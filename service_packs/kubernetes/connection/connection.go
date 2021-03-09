// Package connection is a wrapper for the connection to the Kubernetes API
package connection

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/citihub/probr/audit"
	"github.com/citihub/probr/config"
	"github.com/citihub/probr/service_packs/kubernetes/errors"
	"github.com/citihub/probr/utils"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/remotecommand"
	executil "k8s.io/client-go/util/exec"
)

// Conn simplifies the kubernetes API connection
type Conn struct {
	clientSet         *kubernetes.Clientset
	clientConfig      *rest.Config
	clusterIsDeployed error
}

// Connection should be used instead of Conn within probes to allow mocking during testing
type Connection interface {
	ClusterIsDeployed() error
	CreatePodFromObject(pod *apiv1.Pod, probeName string) (*apiv1.Pod, error)
	DeletePodIfExists(podName, namespace, probeName string) error
	ExecCommand(command, namespace, podName string) (status int, stdout string, err error)
	GetRawResourcesByApiEndpoint(apiEndPoint string) (*K8SJSON, error)
}

// K8SJSON encapsulates the response from a raw/rest call to the Kubernetes API
type K8SJSON struct {
	APIVersion string
	Items      []struct {
		Kind     string
		Metadata map[string]string
	}
}

var instance *Conn
var once sync.Once

// Get retrieves the connection object. Instantiates the connection if necessary
func Get() *Conn {
	once.Do(func() {
		instance = &Conn{}
		instance.setClientConfig()
		instance.setClientSet()
		instance.bootstrapDefaultNamespace()
	})
	return instance
}

// ClusterIsDeployed verifies that the connection instantiation did not report a failure at any point
func (connection *Conn) ClusterIsDeployed() error {
	return connection.clusterIsDeployed
}

func (connection *Conn) setClientSet() {
	var err error
	connection.clientSet, err = kubernetes.NewForConfig(connection.clientConfig)
	if err != nil {
		connection.clusterIsDeployed = utils.ReformatError("Failed to create Kubernetes client set: %v", err)
	}
}

// GetOrCreateNamespace will retrieve or create a namespace within the current Kubernetes cluster
func (connection *Conn) GetOrCreateNamespace(namespace string) (*apiv1.Namespace, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	namespaceObject := apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	createdNamespace, err := connection.clientSet.CoreV1().Namespaces().Create(
		ctx, &namespaceObject, metav1.CreateOptions{})

	if err != nil {
		if errors.IsStatusCode(409, err) {
			log.Printf("[INFO] Namespace %v already exists. Returning existing.", namespace)
			//return it and nil out the err
			return createdNamespace, nil
		}
		return nil, err
	}

	log.Printf("[INFO] Namespace %q created.", createdNamespace.GetObjectMeta().GetName())

	return createdNamespace, nil
}

// CreatePodFromObject creates a pod from the supplied pod object within an existing namespace
func (connection *Conn) CreatePodFromObject(pod *apiv1.Pod, probeName string) (*apiv1.Pod, error) {
	podName := pod.ObjectMeta.Name
	namespace := pod.ObjectMeta.Namespace

	if pod == nil || podName == "" || namespace == "" {
		return nil, fmt.Errorf("one or more of pod (%v), podName (%v) or namespace (%v) is nil - cannot create POD", pod, podName, namespace)
	}

	log.Printf("[INFO] Creating pod %v in namespace %v", podName, namespace)
	log.Printf("[DEBUG] Pod details: %+v", *pod)

	c := connection.clientSet

	podsClient := c.CoreV1().Pods(namespace)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := podsClient.Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		log.Printf("[INFO] Attempt to create pod '%v' failed with error: '%v'", podName, err)
	} else {
		log.Printf("[INFO] Attempt to create pod '%v' succeeded", podName)
		audit.State.GetProbeLog(probeName).CountPodCreated(podName)
	}
	return res, err
}

// DeletePodIfExists deletes the given pod in the specified namespace.
func (connection *Conn) DeletePodIfExists(podName, namespace, probeName string) error {
	clientSet, _ := kubernetes.NewForConfig(connection.clientConfig)
	podsClient := clientSet.CoreV1().Pods(namespace)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("[DEBUG] Attempting to delete pod: %s", podName)

	err := podsClient.Delete(ctx, podName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	audit.State.GetProbeLog(probeName).CountPodDestroyed()
	log.Printf("[INFO] POD %s deleted.", podName)
	return nil
}

// ExecCommand executes the supplied command on the given pod name in the specified namespace.
func (connection *Conn) ExecCommand(cmd, namespace, podName string) (status int, stdout string, err error) {
	status = -1
	if cmd == "" {
		err = utils.ReformatError("Command string not provided to ExecCommand")
		return
	}
	connection.waitForPod(namespace, podName)

	log.Printf("[DEBUG] Executing command: \"%s\" on POD '%s' in namespace '%s'", cmd, podName, namespace)
	request := connection.clientSet.CoreV1().RESTClient().Post().Resource("pods").
		Name(podName).Namespace(namespace).SubResource("exec")

	scheme := runtime.NewScheme()
	if err = apiv1.AddToScheme(scheme); err != nil {
		err = utils.ReformatError("Could not add to scheme: %v", err)
		return
	}

	parameterCodec := runtime.NewParameterCodec(scheme)
	options := apiv1.PodExecOptions{
		Command: strings.Fields(cmd),
		Stdout:  true,
		Stderr:  true,
		TTY:     false,
		// 'Container:' is required if more than one container exists
	}

	request.VersionedParams(&options, parameterCodec)

	log.Printf("[DEBUG] %s.%s: ExecCommand Request URL: %v", utils.CallerName(2), utils.CallerName(1), request.URL().String())
	config, err := clientcmd.BuildConfigFromFlags("", config.Vars.ServicePacks.Kubernetes.KubeConfigPath)
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", request.URL())
	if err != nil {
		err = utils.ReformatError("Failed to create Executor: %v", err)
		return
	}

	var stdoutBuffer, stderrBuffer bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: &stdoutBuffer,
		Stderr: &stderrBuffer,
		Tty:    false,
	})
	stdout = stdoutBuffer.String()
	if err != nil {
		if exit, ok := err.(executil.CodeExitError); ok {
			//the command has been executed on the container, but the underlying command raised an error
			//this is an 'external' error and represents a successful communication with the cluster
			err = utils.ReformatError(fmt.Sprintf("err: %s ; stdout: %s ; stderr: %s", err, stdout, stderrBuffer.String()))
			status = exit.Code
			return
		}
		// Internal error
		err = utils.ReformatError("Issue in Stream: %v", err)
	}
	return
}

// GetRawResourcesByApiEndpoint makes a 'raw' REST call to k8s to get the resources specified by the
// supplied endpoint, e.g. "apis/aadpodidentity.k8s.io/v1/azureidentitybindings".
// This is used to interact with available custom resources in the cluster, such as azureidentitybindings
func (connection *Conn) GetRawResourcesByApiEndpoint(apiEndPoint string) (*K8SJSON, error) {

	r := connection.clientSet.CoreV1().RESTClient().Get().AbsPath(apiEndPoint)
	log.Printf("[DEBUG] REST request: %+v", r)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res := r.Do(ctx)
	if res.Error() != nil {
		return nil, res.Error()
	}

	b, _ := res.Raw()
	bs := string(b)
	log.Printf("[DEBUG] STRING result: %v", bs)

	j := K8SJSON{}
	json.Unmarshal(b, &j)

	log.Printf("[DEBUG] JSON result: %+v", j)

	return &j, nil
}

func (connection *Conn) setClientConfig() {
	// Adapted from clientcmd.BuildConfigFromFlags:
	// https://github.com/kubernetes/client-go/blob/5ab99756f65dbf324e5adf9bd020a20a024bad85/tools/clientcmd/client_config.go#L606
	var err error
	vars := &config.Vars.ServicePacks.Kubernetes

	configLoader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(

		&clientcmd.ClientConfigLoadingRules{ExplicitPath: vars.KubeConfigPath},
		&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: ""}})
	rawConfig, _ := configLoader.RawConfig()

	if vars.KubeContext == "" {
		log.Printf("[INFO] Initializing client with default context")
	} else {
		log.Printf("[INFO] Initializing client with context specified in config vars: %v", vars.KubeContext)
		connection.modifyContext(rawConfig, vars.KubeContext)
	}

	connection.clientConfig, err = configLoader.ClientConfig()
	if err != nil {
		connection.clusterIsDeployed = utils.ReformatError("Failed to retrieve rest client config to validate cluster: %v", err)
	}
}

func (connection *Conn) bootstrapDefaultNamespace() {
	_, err := connection.GetOrCreateNamespace(config.Vars.ServicePacks.Kubernetes.ProbeNamespace)
	if err != nil {
		connection.clusterIsDeployed = utils.ReformatError("Failed to retrieve or create default Probr namespace: %v", err)
	}
}

func (connection *Conn) modifyContext(rawConfig clientcmdapi.Config, context string) {
	log.Printf("[DEBUG] Modifying Kubernetes context based on Probr config vars")
	if rawConfig.Contexts[context] == nil {
		connection.clusterIsDeployed = utils.ReformatError("Required context does not exist in provided kubeconfig: %v", context)
	}
	rawConfig.CurrentContext = context
	err := clientcmd.ModifyConfig(clientcmd.NewDefaultPathOptions(), rawConfig, true)
	if err != nil {
		connection.clusterIsDeployed = utils.ReformatError("Failed to modify context in kubeconfig: %v", context)
	}
}

// waitForPod ensures pod has entered a running state, or returns any error encountered
func (connection *Conn) waitForPod(namespace string, podName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	ps := connection.clientSet.CoreV1().Pods(namespace)
	w, err := ps.Watch(ctx, metav1.ListOptions{})

	if err != nil {
		return err
	}

	log.Printf("[INFO] *** Waiting for pod: %s", podName)
	for e := range w.ResultChan() {
		log.Printf("[DEBUG] Watch Probe Type: %v", e.Type)
		pod, ok := e.Object.(*apiv1.Pod)
		if !ok {
			log.Printf("[WARN] Unexpected Watch Probe Type - skipping")
			if ctx.Err() != nil {
				log.Printf("[WARN] Context error received while waiting on pod %v. Error: %v", podName, ctx.Err())
				return ctx.Err()
			}
			continue
		}
		if pod.GetName() != podName {
			continue
		}

		log.Printf("[INFO] Pod %v Phase: %v", pod.GetName(), pod.Status.Phase)
		for _, con := range pod.Status.ContainerStatuses {
			log.Printf("[DEBUG] Container Status: %+v", con)
		}

		err := connection.podInErrorState(pod)
		if err != nil {
			return err
		}

		if pod.Status.Phase == apiv1.PodRunning {
			break
		}

	}
	return nil
}

func (connection *Conn) podInErrorState(p *apiv1.Pod) error {
	if len(p.Status.ContainerStatuses) > 0 {
		if p.Status.ContainerStatuses[0].State.Waiting != nil {
			podName := p.GetObjectMeta().GetName()
			waitReason := p.Status.ContainerStatuses[0].State.Waiting.Reason
			log.Printf("[DEBUG] Pod: %v Waiting reason: %v", podName, waitReason)

			if strings.Contains(waitReason, "error") {
				return utils.ReformatError("Pod '%s' is in an error state: %v", podName, waitReason)
			}
		}
	}
	return nil
}
