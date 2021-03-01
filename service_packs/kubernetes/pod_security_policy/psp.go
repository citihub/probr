package psp

import (
	"fmt"
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
	probeAudit *audit.Probe
	audit      *audit.ScenarioAudit
	podStates  []kubernetes.PodState
}

// Probe meets the service pack interface for adding the logic from this file
var Probe probeStruct

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
func (scenario *scenarioState) podCreationSuccedsWithXSetToYInThePodSpec(key, value string) error {
	// Supported keys:
	//    'allowPrivilegeEscalation'
	//
	// Supported values:
	//    'true'
	//    'false'
	//    'not have a value provided'

	var boolValue, useValue bool
	stepTrace, payload, err := utils.AuditPlaceholders()
	defer func() {
		scenario.audit.AuditScenarioStep(stepTrace.String(), payload, err)
	}()

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
	createdPodObject, creationErr := conn.CreatePodFromObject(podObject)

	stepTrace.WriteString(fmt.Sprintf("Validate successful pod creation; "))
	if creationErr != nil {
		err = utils.ReformatError("Pod creation did not succeed: %v", creationErr)
	} else {
		scenario.probeAudit.CountPodCreated(createdPodObject.ObjectMeta.Name)
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
func (scenario *scenarioState) podCreationFailsWhenXIsSetToYDueToZ(key, value string) error {
	// TODO: Attempt to create a pod from a YAML spec
	// with key/value parameters parsed into it
	// and expect a failure with provided message

	// And pod creation fails when "allowPrivilegeEscalation" is set to "true" in the pod spec due to "restrictions in requesting privileged access"

	// Supported keys:
	//    'allowPrivilegeEscalation'
	//
	// Supported values:
	//    'true'
	//    'false'
	//    'not have a value provided'

	var boolValue, useValue bool
	stepTrace, payload, err := utils.AuditPlaceholders()
	defer func() {
		scenario.audit.AuditScenarioStep(stepTrace.String(), payload, err)
	}()

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
	createdPodObject, creationErr := conn.CreatePodFromObject(podObject)

	stepTrace.WriteString(fmt.Sprintf("Confirm pod creation failure; "))
	if creationErr == nil {
		err = utils.ReformatError("Pod creation was sucessful, while expecting error") //TODO: Reword this
		scenario.probeAudit.CountPodCreated(createdPodObject.ObjectMeta.Name)
	} else {
		if !errors.IsStatusCode403(creationErr) { //Expected failure due to unauthorized error
			err = utils.ReformatError("Unexpected error during Pod creation : %v", creationErr)
		}
	}

	payload = struct {
		RequestedPod  *apiv1.Pod
		CreationError error
	}{
		RequestedPod:  podObject,
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
	ctx.Step(`^pod creation succeeds with "([^"]*)" set to "([^"]*)" in the pod spec$`, scenario.podCreationSuccedsWithXSetToYInThePodSpec)
	ctx.Step(`^pod creation fails when "([^"]*)" is set to "([^"]*)" in the pod spec due to "([^"]*)"$`, scenario.podCreationFailsWhenXIsSetToYDueToZ)
	ctx.Step(`^the execution of a "([^"]*)" command inside the Pod is successful$`, scenario.theExecutionOfAXCommandInsideThePodIsSuccessful)
	ctx.Step(`^the execution of a "([^"]*)" command inside the Pod fails due to "([^"]*)"$`, scenario.theExecutionOfAXCommandInsideThePodFailsDueToY)

	ctx.AfterScenario(func(s *godog.Scenario, err error) {
		// if kubernetes.GetKeepPodsFromConfig() == false {
		// 	for pod := range scenario.podStates {
		// 		connection.DeletePod(pod, kubernetes.Namespace, probe.Name()) // TODO
		// 	}
		// }
		coreengine.LogScenarioEnd(s)
	})
}

func beforeScenario(s *scenarioState, probeName string, gs *godog.Scenario) {
	s.name = gs.Name
	s.probeAudit = audit.State.GetProbeLog(probeName)
	s.audit = audit.State.GetProbeLog(probeName).InitializeAuditor(gs.Name, gs.Tags)
	coreengine.LogScenarioStart(gs)
}
