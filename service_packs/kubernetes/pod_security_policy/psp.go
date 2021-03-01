package psp

import (
	"fmt"
	"log"
	"strconv"

	"github.com/cucumber/godog"

	"github.com/citihub/probr/audit"
	"github.com/citihub/probr/config"
	"github.com/citihub/probr/service_packs/coreengine"
	"github.com/citihub/probr/service_packs/kubernetes"
	"github.com/citihub/probr/service_packs/kubernetes/connection"
	"github.com/citihub/probr/service_packs/kubernetes/constructors"
	"github.com/citihub/probr/service_packs/kubernetes/errors"
	"github.com/citihub/probr/utils"

	apiv1 "k8s.io/api/core/v1"
)

type probeStruct struct {
}

var conn connection.KubernetesAPI

// scenarioState holds the steps and state for any scenario in this probe
type scenarioState struct {
	name       string
	namespace  string
	probeAudit *audit.Probe
	audit      *audit.ScenarioAudit
	pods       []string
}

// Probe meets the service pack interface for adding the logic from this file
var Probe probeStruct

func (scenario *scenarioState) createPodfromObject(podObject *apiv1.Pod) (createdPodObject *apiv1.Pod, err error) {
	createdPodObject, err = conn.CreatePodFromObject(podObject, Probe.Name())
	if err == nil {
		scenario.pods = append(scenario.pods, createdPodObject.ObjectMeta.Name)
	}
	return
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

	return conn.ClusterIsDeployed()
}

// Attempt to deploy a pod from a default pod spec, with specified modification
func (scenario *scenarioState) podCreationResultsWithXSetToYInThePodSpec(result, key, value string) error {
	// Possible Results:
	// 'succeeds'
	// 'fails'
	//
	// Supported keys:
	//    'allowPrivilegeEscalation'
	//
	// Supported values:
	//    'true'
	//    'false'
	//    'not have a value provided'

	stepTrace, payload, err := utils.AuditPlaceholders()
	defer func() {
		scenario.audit.AuditScenarioStep(stepTrace.String(), payload, err)
	}()
	var boolValue, useValue, shouldCreate bool

	switch result {
	case "succeeds":
		shouldCreate = true
	case "fails":
		shouldCreate = false
	default:
		return utils.ReformatError("Unexpected value provided for expected pod creation result: %s", result) // No payload is necessary if an invalid value was provided
	}

	if value != "not have a value provided" {
		useValue = true
		boolValue, err = strconv.ParseBool(value)
		if err != nil {
			return utils.ReformatError("Expected 'true' or 'false' but found '%s'", value) // No payload is necessary if an invalid value was provided
		}
	}

	stepTrace.WriteString(fmt.Sprintf("Build a pod spec with default values; "))
	securityContext := constructors.DefaultContainerSecurityContext()
	podObject := constructors.PodSpec(Probe.Name(), config.Vars.ServicePacks.Kubernetes.ProbeNamespace, securityContext)
	//TODO: Unit test that this always is true: len(podObject.Spec.Containers) > 0

	if useValue {
		stepTrace.WriteString(fmt.Sprintf("Set '%v' to '%v' in pod spec; ", key, value))
		switch key {
		case "allowPrivilegeEscalation":
			podObject.Spec.Containers[0].SecurityContext.AllowPrivilegeEscalation = &boolValue
		default:
			return utils.ReformatError("Unsupported key provided: %s", key) // No payload is necessary if an invalid key was provided
		}
	}

	stepTrace.WriteString(fmt.Sprintf("Create pod from spec; "))
	createdPodObject, creationErr := scenario.createPodfromObject(podObject)

	stepTrace.WriteString(fmt.Sprintf("Validate pod creation %s; ", result))

	// Leaving these checks verbose for clarity
	if shouldCreate {
		if creationErr != nil {
			err = utils.ReformatError("Pod creation did not succeed: %v", creationErr)
		}
	} else if !shouldCreate {
		if creationErr == nil {
			err = utils.ReformatError("Pod creation succeeded, but should have failed")
		} else {
			if !errors.IsStatusCode403(creationErr) {
				err = utils.ReformatError("Unexpected error during Pod creation : %v", creationErr)
			}
		}
	}

	payload = struct {
		RequestedPod  *apiv1.Pod
		CreatedPod    *apiv1.Pod
		CreationError error
	}{
		RequestedPod:  podObject,
		CreatedPod:    createdPodObject,
		CreationError: creationErr,
	}

	return err
}

func (scenario *scenarioState) theExecutionOfAXCommandInsideThePodIsSuccessful(permission string) error {
	// permission = 'non-privileged' / 'privileged'
	// TODO: Attempt to execute a command on the pod in podStates
	// and expect an exit code of zero
	return nil
}

func (scenario *scenarioState) theExecutionOfAXCommandInsideThePodFailsDueToY(permission, reason string) error {
	// permission = 'non-privileged' / 'privileged'
	// TODO: Attempt to execute a command on the pod in podStates
	// and expect a non-zero exit code with an error that equates to the specified reason
	return nil
}

// Name presents the name of this probe for external reference
func (probe probeStruct) Name() string {
	return "pod_security_policy"
}

// Path presents the path of these feature files for external reference
func (probe probeStruct) Path() string {
	return coreengine.GetFeaturePath("service_packs", "kubernetes", probe.Name())
}

// ProbeInitialize handles any overall Test Suite initialisation steps.  This is registered with the
// test handler as part of the init() function.
func (probe probeStruct) ProbeInitialize(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
		conn = connection.Get()
	})

	ctx.AfterSuite(func() {
	})
}

// ScenarioInitialize initialises the specific test steps.  This is essentially the creation of the test
// which reflects the tests described in the events directory.  There must be a test step registered for
// each line in the feature files. Note: Godog will output stub steps and implementations if it doesn't find
// a step / function defined.  See: https://github.com/cucumber/godog#example.
func (probe probeStruct) ScenarioInitialize(ctx *godog.ScenarioContext) {
	scenario := scenarioState{}

	ctx.BeforeScenario(func(s *godog.Scenario) {
		beforeScenario(&scenario, probe.Name(), s)
	})

	// Background
	ctx.Step(`^a Kubernetes cluster exists which we can deploy into$`, scenario.aKubernetesClusterIsDeployed)

	// Scenarios
	ctx.Step(`^pod creation "([^"]*)" with "([^"]*)" set to "([^"]*)" in the pod spec$`, scenario.podCreationResultsWithXSetToYInThePodSpec)
	ctx.Step(`^pod creation "([^"]*)" with "([^"]*)" set to "([^"]*)" in the pod spec$`, scenario.podCreationResultsWithXSetToYInThePodSpec)
	ctx.Step(`^the execution of a "([^"]*)" command inside the Pod is successful$`, scenario.theExecutionOfAXCommandInsideThePodIsSuccessful)
	ctx.Step(`^the execution of a "([^"]*)" command inside the Pod fails due to "([^"]*)"$`, scenario.theExecutionOfAXCommandInsideThePodFailsDueToY)

	ctx.AfterScenario(func(s *godog.Scenario, err error) {
		if kubernetes.GetKeepPodsFromConfig() == false {
			for _, podName := range scenario.pods {
				err = conn.DeletePodIfExists(podName, scenario.namespace, probe.Name())
				if err != nil {
					log.Printf(fmt.Sprintf("[ERROR] Could not retrieve pod for deletion: %s", err))
				}
			}
		}
		coreengine.LogScenarioEnd(s)
	})
}

func beforeScenario(s *scenarioState, probeName string, gs *godog.Scenario) {
	s.name = gs.Name
	s.probeAudit = audit.State.GetProbeLog(probeName)
	s.audit = audit.State.GetProbeLog(probeName).InitializeAuditor(gs.Name, gs.Tags)
	s.pods = make([]string, 0)
	s.namespace = config.Vars.ServicePacks.Kubernetes.ProbeNamespace
	coreengine.LogScenarioStart(gs)
}
