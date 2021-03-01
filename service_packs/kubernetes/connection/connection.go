// Package connection is a wrapper for the connection to the Kubernetes API
package connection

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/citihub/probr/config"
	"github.com/citihub/probr/service_packs/kubernetes/errors"
	"github.com/citihub/probr/utils"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// Conn represents the k8s API connection.
type Conn struct {
	// clientSet is private in order to restrict probes from implementing client-go logic directly
	clientSet    *kubernetes.Clientset
	clientConfig *rest.Config
	// clientMutex       sync.Mutex
	clusterIsDeployed error
}

// KubernetesAPI should be used instead of Conn within probes to allow mocking during testing
type KubernetesAPI interface {
	ClusterIsDeployed() error
	CreatePodFromObject(*apiv1.Pod) (*apiv1.Pod, error)
}

var instance *Conn
var once sync.Once

// Get ...
func Get() *Conn {
	once.Do(func() {
		instance = &Conn{}
		instance.setClientConfig()
		instance.setClientSet()
		instance.bootstrapDefaultNamespace()
	})
	return instance
}

// ClusterIsDeployed verifies a cluster using the privided kubernetes config and context.
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
		if errors.IsStatusCode409(err) {
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
func (connection *Conn) CreatePodFromObject(pod *apiv1.Pod) (*apiv1.Pod, error) {
	podName := pod.ObjectMeta.Name
	namespace := pod.ObjectMeta.Namespace

	if pod == nil || podName == "" || namespace == "" {
		return nil, fmt.Errorf("one or more of pod (%v), podName (%v) or namespace (%v) is nil - cannot create POD", pod, podName, namespace)
	}

	log.Printf("[INFO] Creating pod %v in namespace %v", podName, namespace)
	log.Printf("[DEBUG] Pod details: %+v", *pod)

	c := connection.clientSet

	podsMgr := c.CoreV1().Pods(namespace) //TODO: Rename this obj to something more meaningful

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := podsMgr.Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		log.Printf("[INFO] Attempt to create pod '%v' failed with error: '%v'", podName, err)
	} else {
		log.Printf("[INFO] Attempt to create pod '%v' succeeded", podName)
		//probe.CountPodCreated(podName) //TODO: This should be moved up to the probe or pack level. Currentl is not logging.
	}

	// TODO: We are not waiting for PodState to be running here, like it is done in kube object. TBD.

	return res, err
}
