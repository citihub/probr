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
	// TODO: This is only here because it works, we need to revisit it and gain a fuller understanding of the different singleton options
	once.Do(func() {
		instance = &Conn{}
		instance.getClientConfig()
		instance.getClientSet() // After this point we'll know whether the cluster specified in the kubeconfig is deployed and accessible
	})
	return instance
}

// ClusterIsDeployed verifies if a cluster is deployed that can be contacted based on the current
// kubernetes config and context.
func (connection *Conn) ClusterIsDeployed() error {
	return connection.clusterIsDeployed
}

func (connection *Conn) getClientSet() {
	var err error
	connection.clientSet, err = kubernetes.NewForConfig(connection.clientConfig)
	if err != nil {
		connection.clusterIsDeployed = utils.ReformatError("Failed to rest client config: %v", err)
	}
}

func (connection *Conn) getClientConfig() {
	// Adapted from clientcmd.BuildConfigFromFlags:
	// https://github.com/kubernetes/client-go/blob/5ab99756f65dbf324e5adf9bd020a20a024bad85/tools/clientcmd/client_config.go#L606
	var err error
	vars := &config.Vars.ServicePacks.Kubernetes

	configLoader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: vars.KubeConfigPath},
		&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: ""}})
	rawConfig, _ := configLoader.RawConfig()

	if vars.KubeContext == "" {
		log.Printf("[NOTICE] Initializing client with default context from provided config")
	} else {
		log.Printf("[NOTICE] Initializing client with non-default context: %v", vars.KubeContext)
		connection.modifyContext(rawConfig, vars.KubeContext)
	}

	connection.clientConfig, err = configLoader.ClientConfig()
	if err != nil {
		connection.clusterIsDeployed = utils.ReformatError("Failed to rest client config: %v", err)
	}
}

func (connection *Conn) modifyContext(rawConfig clientcmdapi.Config, context string) {
	if rawConfig.Contexts[context] == nil {
		connection.clusterIsDeployed = utils.ReformatError("Required context does not exist in provided kubeconfig: %v", context)
	}
	rawConfig.CurrentContext = context
	err := clientcmd.ModifyConfig(clientcmd.NewDefaultPathOptions(), rawConfig, true)
	if err != nil {
		connection.clusterIsDeployed = utils.ReformatError("Failed to modify context in kubeconfig: %v", context)
	}
}

func (connection *Conn) getOrCreateNamespace(ns *string) (*apiv1.Namespace, error) {

	c := connection.clientSet

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//try and create ...
	apiNS := apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: *ns,
		},
	}
	n, err := c.CoreV1().Namespaces().Create(ctx, &apiNS, metav1.CreateOptions{})
	if err != nil {
		if errors.IsStatusCode409(err) {
			log.Printf("[INFO] Namespace %v already exists. Returning existing.", *ns)
			//return it and nil out the err
			return n, nil
		}
		return nil, err
	}

	log.Printf("[INFO] Namespace %q created.", n.GetObjectMeta().GetName())

	return n, nil
}

// CreatePodFromObject creates a pod from the supplied pod object with the given pod name and namespace.
func (connection *Conn) CreatePodFromObject(pod *apiv1.Pod) (*apiv1.Pod, error) {
	podName := pod.ObjectMeta.Name
	namespace := pod.ObjectMeta.Namespace

	if pod == nil || podName == "" || namespace == "" {
		return nil, fmt.Errorf("one or more of pod (%v), podName (%v) or namespace (%v) is nil - cannot create POD", pod, podName, namespace)
	}

	log.Printf("[INFO] Creating pod %v in namespace %v", podName, namespace)
	log.Printf("[DEBUG] Pod details: %+v", *pod)

	c := connection.clientSet

	// TODO: This looks like bootstrp logic. It should be moved out.
	// 		If specified namespace doesn't exist is better to return error (Single Responsibility Principle)
	//create the namespace for the POD (noOp if already present)
	_, err := connection.getOrCreateNamespace(&namespace)
	if err != nil {
		return nil, err
	}

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
