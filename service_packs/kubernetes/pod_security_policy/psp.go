package psp

import (
	"fmt"
	"log"
	"strconv"
	"strings"

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

var conn connection.Connection

// scenarioState holds the steps and state for any scenario in this probe
type scenarioState struct {
	name        string
	currentStep string
	namespace   string
	probeAudit  *audit.Probe
	audit       *audit.ScenarioAudit
	pods        []string
	given       bool
}

// Probe meets the service pack interface for adding the logic from this file
var Probe probeStruct
var scenario scenarioState

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
		scenario.audit.AuditScenarioStep(scenario.currentStep, stepTrace.String(), payload, err)
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

func (scenario *scenarioState) toDo(todo string) error {
	stepTrace, payload, err := utils.AuditPlaceholders()
	defer func() {
		scenario.audit.AuditScenarioStep(scenario.currentStep, stepTrace.String(), payload, err)
	}()
	stepTrace.WriteString(fmt.Sprintf("This step was included to inform developers that a scenario is incomplete; "))
	payload = struct {
		TODO string
	}{TODO: todo}
	return godog.ErrPending
}

// Attempt to deploy a pod from a default pod spec, with specified modification
func (scenario *scenarioState) podCreationResultsWithXSetToYInThePodSpec(result, key, value string) error {
	// Supported results:
	//     'succeeds'
	//     'fails'
	//
	// Supported keys:
	//    'allowPrivilegeEscalation'
	//    'hostPID'
	//    'hostIPC'
	//
	// Supported values:
	//    'true'
	//    'false'
	//    'not have a value provided'

	stepTrace, payload, err := utils.AuditPlaceholders()
	defer func() {
		scenario.audit.AuditScenarioStep(scenario.currentStep, stepTrace.String(), payload, err)
	}()
	var boolValue, useValue, shouldCreate bool

	switch result {
	case "succeeds":
		shouldCreate = true
	case "fails":
		shouldCreate = false
	default:
		err = utils.ReformatError("Unexpected value provided for expected pod creation result: %s", result) // No payload is necessary if an invalid value was provided
		return err
	}

	if value != "not have a value provided" {
		useValue = true
		boolValue, err = strconv.ParseBool(value)
		if err != nil {
			err = utils.ReformatError("Expected 'true' or 'false' but found '%s'", value) // No payload is necessary if an invalid value was provided
			return err
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
		case "hostPID":
			podObject.Spec.HostPID = boolValue
		case "hostIPC":
			podObject.Spec.HostIPC = boolValue
		default:
			err = utils.ReformatError("Unsupported key provided: %s", key) // No payload is necessary if an invalid key was provided
			return err
		}
	}

	stepTrace.WriteString(fmt.Sprintf("Create pod from spec; "))
	createdPodObject, creationErr := scenario.createPodfromObject(podObject)

	stepTrace.WriteString(fmt.Sprintf("Validate pod creation %s; ", result))

	// Leaving these checks verbose for clarity
	switch shouldCreate {
	case true:
		if creationErr != nil {
			err = utils.ReformatError("Pod creation did not succeed: %v", creationErr)
		}
	case false:
		if creationErr == nil {
			err = utils.ReformatError("Pod creation succeeded, but should have failed")
		} else {
			if !errors.IsStatusCode(403, creationErr) {
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

func (scenario *scenarioState) theExecutionOfAXCommandInsideThePodIsY(permission, result string) error {
	// Supported permissions:
	//     'non-privileged'
	//     'privileged'
	//
	// Supported results:
	//     'successful'
	//     'rejected'

	stepTrace, payload, err := utils.AuditPlaceholders()
	defer func() {
		scenario.audit.AuditScenarioStep(scenario.currentStep, stepTrace.String(), payload, err)
	}()

	// Guard clause
	if len(scenario.pods) == 0 {
		err = utils.ReformatError("Pod failed to create in the previous step")
		return err
	}

	var cmd string
	switch permission {
	case "non-privileged":
		cmd = "ls"
	case "privileged":
		cmd = "sudo ls"
	default:
		err = utils.ReformatError("Unexpected value provided for command permission type: %s", permission) // No payload is necessary if an invalid value was provided
		return err
	}

	var expectedExitCode int
	switch result {
	case "successful":
		expectedExitCode = 0
	case "rejected":
		expectedExitCode = 126 // If a command is found but is not executable, the return status is 126
		// Known issue: we can't guarantee that the 126 recieved by kubectl isn't a masked 127
	default:
		err = utils.ReformatError("Unexpected value provided for expected command result: %s", result) // No payload is necessary if an invalid value was provided
		return err

	}
	stepTrace.WriteString("Attempt to run a command in the pod that was created by the previous step; ")
	exitCode, _, err := conn.ExecCommand(cmd, scenario.namespace, scenario.pods[0])

	payload = struct {
		Command          string
		ExitCode         int
		ExpectedExitCode int
	}{
		Command:          cmd,
		ExitCode:         exitCode,
		ExpectedExitCode: expectedExitCode,
	}

	if exitCode == expectedExitCode {
		err = nil
	}
	return err
}

func (scenario *scenarioState) theCommandXShouldOnlyShowTheContainerProcesses(command string) (err error) {
	// Supported commands:
	//     'ps'
	//     'lsns -n'

	stepTrace, payload, err := utils.AuditPlaceholders()
	defer func() {
		scenario.audit.AuditScenarioStep(scenario.currentStep, stepTrace.String(), payload, err)
	}()
	exitCode, stdout, err := conn.ExecCommand(command, scenario.namespace, scenario.pods[0])

	entrypoint := strings.Join(constructors.DefaultEntrypoint(), " ")

	// NOTE: This expectation depends on using DefaultPodSecurityContext during the previous step
	switch command {
	case "ps":
		stepTrace.WriteString("Validating that the container's entrypoint is PID 1 in the process tree; ")
		expected := fmt.Sprintf("1 1000      0:00 %s", entrypoint)
		if !strings.Contains(stdout, expected) {
			err = utils.ReformatError("Expected to find container entrypoint, but did not")
		}
	case "lsns -n":
		stepTrace.WriteString("Validating that no namespace has an entrypoint different from the container's entrypoint; ")
		stdoutLines := strings.Split(stdout, "\n")
		for _, entry := range stdoutLines {
			if entry != "" && !strings.Contains(entry, entrypoint) {
				err = utils.ReformatError("A namespace is visible that uses a different entrypoint from the container, suggesting that hostIPC was used")
			}
		}
	default:
		err = utils.ReformatError("Unsupported value provided for command")
	}

	// TODO: Validate that this fails as expected
	payload = struct {
		Command    string
		ExitCode   int
		Stdout     string
		Entrypoint string
	}{
		Command:    command,
		ExitCode:   exitCode,
		Stdout:     stdout,
		Entrypoint: entrypoint,
	}
	return
}

func (scenario *scenarioState) theHostNamespaceShouldNotBeVisible() (err error) {
	stepTrace, payload, err := utils.AuditPlaceholders()
	defer func() {
		scenario.audit.AuditScenarioStep(scenario.currentStep, stepTrace.String(), payload, err)
	}()

	cmd := "lsns -l"
	exitCode, stdout, err := conn.ExecCommand(cmd, scenario.namespace, scenario.pods[0])
	payload = struct {
		Command  string
		ExitCode int
		Stdout   string
	}{
		Command:  cmd,
		ExitCode: exitCode,
		Stdout:   stdout,
	}
	return
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

	ctx.BeforeScenario(func(s *godog.Scenario) {
		beforeScenario(&scenario, probe.Name(), s)
	})

	// Background
	ctx.Step(`^a Kubernetes cluster exists which we can deploy into$`, scenario.aKubernetesClusterIsDeployed)

	// Use for steps that have yet to be written
	ctx.Step(`^TODO: "([^"]*)"$`, scenario.toDo)

	// Parameterized Scenarios
	ctx.Step(`^pod creation "([^"]*)" with "([^"]*)" set to "([^"]*)" in the pod spec$`, scenario.podCreationResultsWithXSetToYInThePodSpec)
	ctx.Step(`^pod creation "([^"]*)" with "([^"]*)" set to "([^"]*)" in the pod spec$`, scenario.podCreationResultsWithXSetToYInThePodSpec)
	ctx.Step(`^the execution of a "([^"]*)" command inside the Pod is "([^"]*)"$`, scenario.theExecutionOfAXCommandInsideThePodIsY)
	ctx.Step(`^the command "([^"]*)" should only show the container processes$`, scenario.theCommandXShouldOnlyShowTheContainerProcesses)

	ctx.AfterScenario(func(s *godog.Scenario, err error) {
		afterScenario(scenario, probe, s, err)
	})

	ctx.BeforeStep(func(st *godog.Step) {
		scenario.currentStep = st.Text
	})

	ctx.AfterStep(func(st *godog.Step, err error) {
		scenario.currentStep = ""
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

func afterScenario(scenario scenarioState, probe probeStruct, gs *godog.Scenario, err error) {
	if kubernetes.GetKeepPodsFromConfig() == false {
		for _, podName := range scenario.pods {
			err = conn.DeletePodIfExists(podName, scenario.namespace, probe.Name())
			if err != nil {
				log.Printf(fmt.Sprintf("[ERROR] Could not retrieve pod from namespace '%s' for deletion: %s", scenario.namespace, err))
			}
		}
	}
	coreengine.LogScenarioEnd(gs)
}
