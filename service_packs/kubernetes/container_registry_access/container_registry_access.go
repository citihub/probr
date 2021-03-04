// Package cra provides the implementation required to execute the
// feature based test cases described in the the 'events' directory. //TODO: Clarify what 'events' directory is
package cra

import (
	"fmt"
	"log"

	"github.com/cucumber/godog"
	apiv1 "k8s.io/api/core/v1"

	"github.com/citihub/probr/audit"
	"github.com/citihub/probr/config"
	"github.com/citihub/probr/service_packs/coreengine"
	"github.com/citihub/probr/service_packs/kubernetes"
	"github.com/citihub/probr/service_packs/kubernetes/connection"
	"github.com/citihub/probr/service_packs/kubernetes/constructors"
	"github.com/citihub/probr/service_packs/kubernetes/errors"
	"github.com/citihub/probr/utils"
)

type probeStruct struct{}

// Will provide functionality to interact with K8s cluster
var conn connection.Connection

// scenarioState holds the steps and state for any scenario in this probe
type scenarioState struct {
	name             string
	namespace        string
	audit            *audit.ScenarioAudit
	probe            *audit.Probe
	podState         kubernetes.PodState // TODO: Remove?
	pods             []string
	podCreationError error
}

// Probe meets the service pack interface for adding the logic from this file
var Probe probeStruct
var scenario scenarioState

// ContainerRegistryAccess is the section of the kubernetes package which provides the kubernetes interactions required to support
// container registry scenarios.
var cra ContainerRegistryAccess //TODO: Remove?

//TODO: Remove?
// SetContainerRegistryAccess allows injection of ContainerRegistryAccess helper.
func SetContainerRegistryAccess(c ContainerRegistryAccess) {
	cra = c
}

func (scenario *scenarioState) aKubernetesClusterIsDeployed() error {
	// Standard auditing logic to ensures panics are also audited
	stepTrace, payload, err := utils.AuditPlaceholders()
	defer func() {
		scenario.audit.AuditScenarioStep(stepTrace.String(), payload, err)
	}()
	stepTrace.WriteString(fmt.Sprintf("Validate that a cluster can be reached using the specified kube config and context; "))

	payload = struct {
		KubeConfigPath string
		KubeContext    string
	}{
		config.Vars.ServicePacks.Kubernetes.KubeConfigPath,
		config.Vars.ServicePacks.Kubernetes.KubeContext,
	}

	err = conn.ClusterIsDeployed() // Must be assigned to 'err' be audited
	return err
}

// // CIS-6.1.3
// // Minimize cluster access to read-only
// func (s *scenarioState) iAmAuthorisedToPullFromAContainerRegistry() error {
// 	// Standard auditing logic to ensures panics are also audited
// 	stepTrace, payload, err := utils.AuditPlaceholders()
// 	defer func() {
// 		s.audit.AuditScenarioStep(stepTrace.String(), payload, err)
// 	}()

// 	// TODO: We are assuming too much here- if the image successfully pulls but fails to build, this will still fail
// 	pod, podAudit, err := cra.SetupContainerAccessProbePod(config.Vars.ServicePacks.Kubernetes.AuthorisedContainerRegistry, s.probe)
// 	err = kubernetes.ProcessPodCreationResult(&s.podState, pod, kubernetes.PSPContainerAllowedImages, err)

// 	stepTrace.WriteString("Attempted to create a new pod using an image pulled from authorized registry; ")
// 	payload = struct {
// 		AuthorizedRegistry string
// 		PodAudit           *kubernetes.PodAudit
// 	}{
// 		AuthorizedRegistry: config.Vars.ServicePacks.Kubernetes.AuthorisedContainerRegistry,
// 		PodAudit:           podAudit,
// 	}
// 	return err
// }

func (scenario *scenarioState) aUserAttemptsToDeployAContainerFromAn_X_Registry(accessLevel string) error {
	// Supported values for 'accessLevel':
	//    'authorized'
	//    'unauthorized'

	// Standard auditing logic to ensures panics are also audited
	stepTrace, payload, err := utils.AuditPlaceholders()
	defer func() {
		scenario.audit.AuditScenarioStep(stepTrace.String(), payload, err)
	}()

	var isRegistryAuthorized bool

	// Validate input values
	switch accessLevel {
	case "authorized":
		isRegistryAuthorized = true
	case "unauthorized":
		isRegistryAuthorized = false
	default:
		err = utils.ReformatError("Unexpected value provided for accessLevel: '%s' Expected values: ['authorized', 'unauthorized']", accessLevel) // No payload is necessary if an invalid value was provided
		return err
	}

	stepTrace.WriteString(fmt.Sprintf("Get appropriate container image from an '%s' registry ; ", accessLevel))
	imageRegistry := getImageFromConfig(isRegistryAuthorized)

	stepTrace.WriteString(fmt.Sprintf("Build a pod spec with default values; "))
	securityContext := constructors.DefaultContainerSecurityContext()
	podObject := constructors.PodSpec(Probe.Name(), scenario.namespace, securityContext)

	stepTrace.WriteString(fmt.Sprintf("Set container image registry to appropriate value in pod spec; "))
	podObject.Spec.Containers[0].Image = imageRegistry

	stepTrace.WriteString(fmt.Sprintf("Create pod from spec; "))
	createdPodObject, creationErr := scenario.createPodfromObject(podObject) // Pod name is saved to scenario state if successful

	payload = struct {
		AccessLevel   string
		ImageRegistry string
		RequestedPod  *apiv1.Pod
		CreatedPod    *apiv1.Pod
		CreationError error
	}{
		AccessLevel:   accessLevel,
		ImageRegistry: imageRegistry,
		RequestedPod:  podObject,
		CreatedPod:    createdPodObject,
		CreationError: creationErr,
	}

	return err
}

// func (s *scenarioState) theDeploymentAttemptIsAllowed() error {
// 	// Standard auditing logic to ensures panics are also audited
// 	stepTrace, payload, err := utils.AuditPlaceholders()
// 	defer func() {
// 		s.audit.AuditScenarioStep(stepTrace.String(), payload, err)
// 	}()

// 	// TODO: Extending the comment in iAmAuthorisedToPullFromAContainerRegistry...
// 	//       This step doesn't validate the attempt being allowed, it validates the success of the deployment
// 	err = kubernetes.AssertResult(&s.podState, "allowed", "")
// 	stepTrace.WriteString("Asserts pod creation result in scenario state is successful; ")
// 	payload = struct {
// 		PodState kubernetes.PodState
// 	}{s.podState}

// 	return err
// }

func (s *scenarioState) theDeploymentAttemptIs_X(permission string) error {

	// Supported values for 'permission':
	//    'allowed'
	//    'denied'

	// Standard auditing logic to ensures panics are also audited
	stepTrace, payload, err := utils.AuditPlaceholders()
	defer func() {
		scenario.audit.AuditScenarioStep(stepTrace.String(), payload, err)
	}()

	var isDeploymentAllowed bool

	stepTrace.WriteString(fmt.Sprintf("Check that pod creation was '%s' in previous step; ", permission))

	// Validate input values
	switch permission {
	case "allowed":
		isDeploymentAllowed = true
	case "denied":
		isDeploymentAllowed = false
	default:
		err = utils.ReformatError("Unexpected value provided for permission: '%s' Expected values: ['allowed', 'denied']", permission) // No payload is necessary if an invalid value was provided
		return err
	}

	switch isDeploymentAllowed {
	case true:
		if !(len(scenario.pods) > 0) { // Check that at least one pod was created in previous step
			err = utils.ReformatError("Pod creation did not succeed in previous step")
		}
	case false:
		// Check that no pods were created in previous step
		if scenario.podCreationError == nil {
			err = utils.ReformatError("Pod creation succeeded, but should have failed")
		} else {
			stepTrace.WriteString(fmt.Sprintf("Check that pod creation failed due to expected reason (403 Forbidden); "))
			if !errors.IsStatusCode(403, scenario.podCreationError) {
				err = utils.ReformatError("Unexpected error during Pod creation: %v", scenario.podCreationError)
			}
		}
	}

	payload = struct {
		CreatedPods   []string
		CreationError error
	}{
		CreatedPods:   scenario.pods,
		CreationError: scenario.podCreationError,
	}

	return err
}

// // CIS-6.1.5
// // Ensure deployment from authorised container registries is allowed
// func (s *scenarioState) aUserAttemptsToDeployUnauthorisedContainer() error {
// 	// Standard auditing logic to ensures panics are also audited
// 	stepTrace, payload, err := utils.AuditPlaceholders()
// 	defer func() {
// 		s.audit.AuditScenarioStep(stepTrace.String(), payload, err)
// 	}()

// 	pod, podAudit, err := cra.SetupContainerAccessProbePod(config.Vars.ServicePacks.Kubernetes.UnauthorisedContainerRegistry, s.probe)

// 	err = kubernetes.ProcessPodCreationResult(&s.podState, pod, kubernetes.PSPContainerAllowedImages, err)

// 	stepTrace.WriteString(fmt.Sprintf(
// 		"Attempts to deploy a container from %s. Retains pod creation result in scenario state; ",
// 		config.Vars.ServicePacks.Kubernetes.UnauthorisedContainerRegistry))
// 	payload = struct {
// 		UnauthorizedRegistry string
// 		PodAudit             *kubernetes.PodAudit
// 	}{
// 		UnauthorizedRegistry: config.Vars.ServicePacks.Kubernetes.UnauthorisedContainerRegistry,
// 		PodAudit:             podAudit,
// 	}
// 	return err
// }

// func (s *scenarioState) theDeploymentAttemptIsDenied() error {
// 	// Standard auditing logic to ensures panics are also audited
// 	stepTrace, payload, err := utils.AuditPlaceholders()
// 	defer func() {
// 		s.audit.AuditScenarioStep(stepTrace.String(), payload, err)
// 	}()

// 	err = kubernetes.AssertResult(&s.podState, "denied", "")
// 	stepTrace.WriteString("Asserts pod creation result in scenario state is denied; ")
// 	payload = struct {
// 		PodState kubernetes.PodState
// 	}{s.podState}

// 	return err
// }

// Name presents the name of this probe for external reference
func (p probeStruct) Name() string {
	return "container_registry_access"
}

// Path presents the path of these feature files for external reference
func (p probeStruct) Path() string {
	return coreengine.GetFeaturePath("service_packs", "kubernetes", p.Name())
}

// ProbeInitialize handles any overall Test Suite initialisation steps.  This is registered with the
// test handler as part of the init() function.
func (p probeStruct) ProbeInitialize(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
		conn = connection.Get()
	})

	ctx.AfterSuite(func() {
	})

	// TODO: Remove this?
	//check dependencies ...
	if cra == nil {
		// not been given one so set default
		cra = NewDefaultCRA()
	}
}

// ScenarioInitialize initialises the specific test steps.  This is essentially the creation of the test
// which reflects the tests described in the events directory.  There must be a test step registered for
// each line in the feature files. Note: Godog will output stub steps and implementations if it doesn't find
// a step / function defined.  See: https://github.com/cucumber/godog#example.
func (probe probeStruct) ScenarioInitialize(ctx *godog.ScenarioContext) {
	//ps := scenarioState{}

	ctx.BeforeScenario(func(s *godog.Scenario) {
		beforeScenario(&scenario, probe.Name(), s)
	})

	// //common
	// ctx.Step(`^a Kubernetes cluster is deployed$`, ps.aKubernetesClusterIsDeployed)

	// Background
	ctx.Step(`^a Kubernetes cluster exists which we can deploy into$`, scenario.aKubernetesClusterIsDeployed)

	// Steps
	ctx.Step(`^a user attempts to deploy a container from an "([^"]*)" registry$`, scenario.aUserAttemptsToDeployAContainerFromAn_X_Registry)
	ctx.Step(`^the deployment attempt is "([^"]*)"$`, scenario.theDeploymentAttemptIs_X)

	// TODO: Clean up

	//CIS-6.1.4
	//ctx.Step(`^a user attempts to deploy a container from an authorised registry$`, scenario.iAmAuthorisedToPullFromAContainerRegistry) // TODO: This step should be modified in the feature file, or a unique function should be written for it

	//CIS-6.1.5
	// ctx.Step(`^a user attempts to deploy a container from an unauthorised registry$`, scenario.aUserAttemptsToDeployUnauthorisedContainer)
	// ctx.Step(`^the deployment attempt is denied$`, scenario.theDeploymentAttemptIsDenied)

	ctx.AfterScenario(func(s *godog.Scenario, err error) {
		afterScenario(scenario, probe, s, err)
	})
}

func beforeScenario(s *scenarioState, probeName string, gs *godog.Scenario) {
	s.name = gs.Name
	s.probe = audit.State.GetProbeLog(probeName)
	s.audit = audit.State.GetProbeLog(probeName).InitializeAuditor(gs.Name, gs.Tags)
	s.pods = make([]string, 0)
	s.namespace = config.Vars.ServicePacks.Kubernetes.ProbeNamespace
	coreengine.LogScenarioStart(gs)
}

func afterScenario(scenario scenarioState, probe probeStruct, gs *godog.Scenario, err error) {
	cra.TeardownContainerAccessProbePod(scenario.podState.PodName, probe.Name()) // TODO: Refactor

	if kubernetes.GetKeepPodsFromConfig() == false { // TODO: Replace kubernetes ?
		for _, podName := range scenario.pods {
			err = conn.DeletePodIfExists(podName, scenario.namespace, probe.Name())
			if err != nil {
				log.Printf(fmt.Sprintf("[ERROR] Could not retrieve pod from namespace '%s' for deletion: %s", scenario.namespace, err))
			}
		}
	}
	coreengine.LogScenarioEnd(gs)
}

func getContainerRegistryFromConfig(accessLevel bool) string {
	if accessLevel {
		return config.Vars.ServicePacks.Kubernetes.AuthorisedContainerRegistry
	}
	return config.Vars.ServicePacks.Kubernetes.UnauthorisedContainerRegistry
}

func getImageFromConfig(accessLevel bool) string {
	registry := getContainerRegistryFromConfig(accessLevel)
	//full image is the repository + the configured image
	return registry + "/" + config.Vars.ServicePacks.Kubernetes.ProbeImage
}

func (scenario *scenarioState) createPodfromObject(podObject *apiv1.Pod) (createdPodObject *apiv1.Pod, err error) {
	createdPodObject, err = conn.CreatePodFromObject(podObject, Probe.Name())
	if err == nil {
		scenario.pods = append(scenario.pods, createdPodObject.ObjectMeta.Name)
	} else {
		scenario.podCreationError = err
	}
	return
}
