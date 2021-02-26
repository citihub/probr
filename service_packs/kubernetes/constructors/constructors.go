package constructors

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/citihub/probr/config"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GenerateUniquePodName creates a unique pod name based on the format: 'baseName'-'nanosecond time'-'random int'.
func GenerateUniquePodName(baseName string) string {
	//take base and add some uniqueness
	t := time.Now()
	rand.Seed(t.UnixNano())
	uniq := fmt.Sprintf("%v-%v", t.Format("020106-150405"), rand.Intn(100))

	return fmt.Sprintf("%v-%v", baseName, uniq)
}

// GetPodSpec constructs a simple pod object using kubernetes API types.
func GetPodSpec(podName string, namespace string, containerName string, securityContext *apiv1.SecurityContext) *apiv1.Pod {

	annotations := make(map[string]string)
	annotations["seccomp.security.alpha.kubernetes.io/pod"] = "runtime/default"

	return &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
			Labels: map[string]string{
				"app": "demo",
			},
			Annotations: annotations,
		},
		Spec: apiv1.PodSpec{
			SecurityContext: GetDefaultPodSecurityContext(),
			Containers: []apiv1.Container{
				{
					Name:            containerName,
					Image:           GetDefaultProbrImageName(),
					ImagePullPolicy: apiv1.PullIfNotPresent,
					Command: []string{
						"sleep",
						"3600",
					},
					SecurityContext: securityContext,
				},
			},
		},
	}
}

func GetDefaultContainerSecurityContext() *apiv1.SecurityContext {
	b := false

	capabilities := apiv1.Capabilities{
		Drop: GetContainerDropCapabilitiesFromConfig(),
	}

	return &apiv1.SecurityContext{
		Privileged:               &b,
		AllowPrivilegeEscalation: &b,
		Capabilities:             &capabilities,
	}
}

func GetDefaultPodSecurityContext() *apiv1.PodSecurityContext {
	var user, grp, fsgrp int64
	user, grp, fsgrp = 1000, 3000, 2000

	return &apiv1.PodSecurityContext{
		RunAsUser:          &user,
		RunAsGroup:         &grp,
		FSGroup:            &fsgrp,
		SupplementalGroups: []int64{1},
	}
}

func GetDefaultProbrNamespace() string {
	return "probr-ns" // TODO: Get from config
}

func GetDefaultProbrContainerName() string {
	return "psp-test" // TODO: Get from config
}

func GetDefaultProbrImageName() string {
	return fmt.Sprintf(
		"%s/%s",
		config.Vars.ServicePacks.Kubernetes.AuthorisedContainerRegistry,
		config.Vars.ServicePacks.Kubernetes.ProbeImage)
}

// GetContainerDropCapabilitiesFromConfig returns Kubernetes.ContainerRequiredDropCapabilities as a list of Capability objects
func GetContainerDropCapabilitiesFromConfig() []apiv1.Capability {
	// Adding all values from config
	dropCapabilitiesFromConfig := config.Vars.ServicePacks.Kubernetes.ContainerRequiredDropCapabilities

	return GetCapabilitiesFromList(dropCapabilitiesFromConfig)
}

// GetCapabilitiesFromList converts a list of strings into a list of capabilities
func GetCapabilitiesFromList(capList []string) []apiv1.Capability {
	var capabilities []apiv1.Capability

	for _, cap := range capList {
		if cap != "" {
			capabilities = append(capabilities, apiv1.Capability(cap))
		}
	}

	return capabilities
}
