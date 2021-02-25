package psp

import (
	"github.com/cucumber/godog"

	"github.com/citihub/probr/audit"
	"github.com/citihub/probr/service_packs/coreengine"
	"github.com/citihub/probr/service_packs/kubernetes"
)

type probeStruct struct{}

// scenarioState holds the steps and state for any scenario in this probe
type scenarioState struct {
	audit     audit.ScenarioAudit
	podStates []kubernetes.PodState
}

// Probe meets the service pack interface for adding the logic from this file
var Probe probeStruct

func (scenario *scenarioState) aKubernetesClusterIsDeployed() error {
	// TODO: Retrieve the configuration for the kubernetes cluster context specified in config.Vars
	return nil
}

func (scenario *scenarioState) podCreationSuccedsWithXSetToYInThePodSpec(key, value string) error {
	// TODO: Attempt to create a pod from a YAML spec
	// with key/value parameters parsed into it
	// expect successful deployment
	return nil
}
func (scenario *scenarioState) podCreationFailsWhenXIsSetToYDueToZ(key, value string) error {
	// TODO: Attempt to create a pod from a YAML spec
	// with key/value parameters parsed into it
	// and expect a failure with provided message
	return nil
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
	return coreengine.GetFeaturePath("service_packs", "kubernetes", p.Name())
}

// ProbeInitialize handles any overall Test Suite initialisation steps.  This is registered with the
// test handler as part of the init() function.
func (probe probeStruct) ProbeInitialize(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
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
	})

	// Background
	ctx.Step(`^a Kubernetes cluster exists which we can deploy into$`, scenario.aKubernetesClusterIsDeployed)

	// Scenarios
	ctx.Step(`^pod creation succeeds with "([^"]*)" set to "([^"]*)" in the pod spec$`, scenario.podCreationSuccedsWithXSetToYInThePodSpec)
	ctx.Step(`^pod creation fails when "([^"]*)" is set to "([^"]*)" in the pod spec due to "([^"]*)"$`, scenario.podCreationFailsWhenXIsSetToYDueToZ)
	ctx.Step(`^the execution of a "([^"]*)" command inside the Pod is successful$`, scenario.theExecutionOfAXCommandInsideThePodIsSuccessful)
	ctx.Step(`^the execution of a "([^"]*)" command inside the Pod fails due to "([^"]*)"$`, scenario.theExecutionOfAXCommandInsideThePodFailsDueToY)

	ctx.AfterScenario(func(s *godog.Scenario, err error) {
		if kubernetes.GetKeepPodsFromConfig() == false {
			if len(scenario.podStates) == 0 {
				//
			} else {
				for _, s := range scenario.podStates {
					//
				}
			}
		}
		coreengine.LogScenarioEnd(s)
	})
}
