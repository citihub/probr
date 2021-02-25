package psp

import (
	"strings"

	"github.com/cucumber/godog"

	"github.com/citihub/probr/service_packs/coreengine"
	"github.com/citihub/probr/service_packs/kubernetes"
)

type probeStruct struct{}

// Probe meets the service pack interface for adding the logic from this file
var Probe probeStruct

func (s *scenarioState) aKubernetesClusterIsDeployed() error {
	var stepTrace strings.Builder
	description, payload, err := kubernetes.ClusterIsDeployed()
	stepTrace.WriteString(description)
	defer func() {
		s.audit.AuditScenarioStep(stepTrace.String(), payload, err)
	}()
	return err //  ClusterIsDeployed will create a fatal error if kubeconfig doesn't validate
}

// Name presents the name of this probe for external reference
func (p probeStruct) Name() string {
	return "pod_security_policy"
}

// Path presents the path of these feature files for external reference
func (p probeStruct) Path() string {
	return coreengine.GetFeaturePath("service_packs", "kubernetes", p.Name())
}

// pspProbeInitialize handles any overall Test Suite initialisation steps.  This is registered with the
// test handler as part of the init() function.
func (p probeStruct) ProbeInitialize(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
	})

	ctx.AfterSuite(func() {
	})
}

// pspScenarioInitialize initialises the specific test steps.  This is essentially the creation of the test
// which reflects the tests described in the events directory.  There must be a test step registered for
// each line in the feature files. Note: Godog will output stub steps and implementations if it doesn't find
// a step / function defined.  See: https://github.com/cucumber/godog#example.
func (p probeStruct) ScenarioInitialize(ctx *godog.ScenarioContext) {
	ps := scenarioState{}

	ctx.BeforeScenario(func(s *godog.Scenario) {
	})

	ctx.Step(`^a Kubernetes cluster exists which we can deploy into$`, ps.aKubernetesClusterIsDeployed)

	ctx.Step(`^a Kubernetes deployment is applied to an existing Kubernetes cluster$`, ps.aKubernetesDeploymentIsAppliedToAnExistingKubernetesCluster)

	ctx.AfterScenario(func(s *godog.Scenario, err error) {
		if kubernetes.GetKeepPodsFromConfig() == false {
			if len(ps.podStates) == 0 {
			} else {
				for _, s := range ps.podStates {
				}
			}
		}
		coreengine.LogScenarioEnd(s)
	})
}
