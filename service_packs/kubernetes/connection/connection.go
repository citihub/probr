// Package connection is a wrapper for the connection to the Kubernetes API
package connection

import (
	"log"
	"sync"

	"github.com/citihub/probr/config"
	"github.com/citihub/probr/utils"
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
	instance.clientSet, err = kubernetes.NewForConfig(connection.clientConfig)
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
