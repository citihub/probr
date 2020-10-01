// Package general provides the implementation required to execute the feature-based test cases
// described in the the 'events' directory.
package general

import (
	"log"
	"strings"

	"github.com/citihub/probr/internal/audit"
	"github.com/citihub/probr/probes"
	"github.com/citihub/probr/probes/kubernetes/probe"

	"github.com/citihub/probr/internal/clouddriver/kubernetes"
	"github.com/citihub/probr/internal/config"
	"github.com/citihub/probr/internal/coreengine"
	"github.com/citihub/probr/internal/utils"
	"github.com/cucumber/godog"
)

type probeState struct {
	name             string
	state            probe.State
	hasWildcardRoles bool
}

const NAME = "general"

// init() registers the feature tests descibed in this package with the test runner (coreengine.TestRunner) via the call
// to coreengine.AddTestHandler.  This links the test - described by the TestDescriptor - with the handler to invoke.  In
// this case, the general test handler is being used (probes.GodogTestHandler) and the GodogTest data provides the data
// require to execute the test.  Specifically, the data includes the Test Suite and Scenario Initializers from this package
// which will be called from probes.GodogTestHandler.  Note: a blank import at probr library level should be done to
// invoke this function automatically on initial load.
func init() {
	td := coreengine.TestDescriptor{Group: coreengine.Kubernetes,
		Category: coreengine.General, Name: NAME}

	coreengine.AddTestHandler(td, &coreengine.GoDogTestTuple{
		Handler: probes.GodogTestHandler,
		Data: &coreengine.GodogTest{
			TestDescriptor:       &td,
			TestSuiteInitializer: TestSuiteInitialize,
			ScenarioInitializer:  ScenarioInitialize,
		},
	})
}

//general
func (p *probeState) aKubernetesClusterIsDeployed() error {
	b := kubernetes.GetKubeInstance().ClusterIsDeployed()

	if b == nil || !*b {
		log.Fatalf("[ERROR] Kubernetes cluster is not deployed")
	}
	e := audit.AuditLog.GetEventLog(NAME)
	e.AuditProbe(p.name, "aKubernetesClusterIsDeployed", nil)
	return nil
}

//@CIS-5.1.3
func (p *probeState) iInspectTheThatAreConfigured(roleLevel string) error {
	var err error
	if roleLevel == "Cluster Roles" {
		l, e := kubernetes.GetKubeInstance().GetClusterRolesByResource("*")
		err = e
		p.hasWildcardRoles = len(*l) > 0

	} else if roleLevel == "Roles" {
		l, e := kubernetes.GetKubeInstance().GetRolesByResource("*")
		err = e
		p.hasWildcardRoles = len(*l) > 0
	}
	if err != nil {
		err = probes.LogAndReturnError("error raised when retrieving roles for rolelevel %v: %v", roleLevel, err)
	}
	event := audit.AuditLog.GetEventLog(NAME)
	event.AuditProbe(p.name, "iInspectTheThatAreConfigured", err)
	return err
}

func (p *probeState) iShouldOnlyFindWildcardsInKnownAndAuthorisedConfigurations() error {
	//we strip out system/known entries in the cluster roles & roles call
	var err error
	if p.hasWildcardRoles {
		err = probes.LogAndReturnError("roles exist with wildcarded resources")
	}
	e := audit.AuditLog.GetEventLog(NAME)
	e.AuditProbe(p.name, "iShouldOnlyFindWildcardsInKnownAndAuthorisedConfigurations", err)

	//good if get to here
	return err
}

//@CIS-5.6.3
func (p *probeState) iAttemptToCreateADeploymentWhichDoesNotHaveASecurityContext() error {
	b := "probr-general"
	n := kubernetes.GenerateUniquePodName(b)
	i := config.Vars.Images.Repository + "/" + config.Vars.Images.BusyBox

	//create pod with nil security context
	pd, err := kubernetes.GetKubeInstance().CreatePod(&n, utils.StringPtr("probr-general-test-ns"), &b, &i, true, nil)

	e := audit.AuditLog.GetEventLog(NAME)
	s := probe.ProcessPodCreationResult(&p.state, pd, kubernetes.UndefinedPodCreationErrorReason, e, err)
	e.AuditProbe(p.name, "iAttemptToCreateADeploymentWhichDoesNotHaveASecurityContext", s)
	return s
}

func (p *probeState) theDeploymentIsRejected() error {
	//looking for a non-nil creation error
	var err error
	if p.state.CreationError == nil {
		err = probes.LogAndReturnError("pod %v was created successfully. Test fail.", p.state.PodName)
	}
	e := audit.AuditLog.GetEventLog(NAME)
	e.AuditProbe(p.name, "theDeploymentIsRejected", err)

	//nil creation error so test pass
	return err
}

//@CIS-6.10.1
// PENDING IMPLEMENTATION
func (p *probeState) iShouldNotBeAbleToAccessTheKubernetesWebUI() error {
	//TODO: will be difficult to test this.  To access it, a proxy needs to be created:
	//az aks browse --resource-group rg-probr-all-policies --name ProbrAllPolicies
	//which will then open a browser at:
	//http://127.0.0.1:8001/api/v1/namespaces/kube-system/services/https:kubernetes-dashboard:/proxy/#/login
	//I don't think this is going to be easy to do from here
	//Is there another test?  Or is it sufficient to verify that no kube-dashboard is running?
	return nil
}

func (p *probeState) theKubernetesWebUIIsDisabled() error {
	//look for the dashboard pod in the kube-system ns
	pl, err := kubernetes.GetKubeInstance().GetPods("kube-system")

	if err != nil {
		err = probes.LogAndReturnError("error raised when trying to retrieve pods %v", err)
	}

	//a "pass" is the abscence of a "kubernetes-dashboard" pod
	for _, p := range pl.Items {
		if strings.HasPrefix(p.Name, "kubernetes-dashboard") {
			err = probes.LogAndReturnError("kubernetes-dashboard pod found (%v) - test fail", p.Name)
			break
		}
	}
	e := audit.AuditLog.GetEventLog(NAME)
	e.AuditProbe(p.name, "theDeploymentIsRejected", err)
	return err
}

func (p *probeState) tearDown() {
	kubernetes.GetKubeInstance().DeletePod(&p.state.PodName, utils.StringPtr("probr-general-test-ns"), false, NAME)
}

// TestSuiteInitialize handles any overall Test Suite initialisation steps.  This is registered with the
// test handler as part of the init() function.
func TestSuiteInitialize(ctx *godog.TestSuiteContext) {

	ctx.BeforeSuite(func() {}) //nothing for now

	ctx.AfterSuite(func() {})

}

// ScenarioInitialize initialises the specific test steps.  This is essentially the creation of the test
// which reflects the tests described in the events directory.  There must be a test step registered for
// each line in the feature files. Note: Godog will output stub steps and implementations if it doesn't find
// a step / function defined.  See: https://github.com/cucumber/godog#example.
func ScenarioInitialize(ctx *godog.ScenarioContext) {
	ps := probeState{}

	ctx.BeforeScenario(func(s *godog.Scenario) {
		ps.name = s.Name
		probes.LogScenarioStart(s)
	})

	//general
	ctx.Step(`^a Kubernetes cluster is deployed$`, ps.aKubernetesClusterIsDeployed)

	//@CIS-5.1.3
	ctx.Step(`^I inspect the "([^"]*)" that are configured$`, ps.iInspectTheThatAreConfigured)
	ctx.Step(`^I should only find wildcards in known and authorised configurations$`, ps.iShouldOnlyFindWildcardsInKnownAndAuthorisedConfigurations)

	//@CIS-5.6.3
	ctx.Step(`^I attempt to create a deployment which does not have a Security Context$`, ps.iAttemptToCreateADeploymentWhichDoesNotHaveASecurityContext)
	ctx.Step(`^the deployment is rejected$`, ps.theDeploymentIsRejected)

	ctx.Step(`^I should not be able to access the Kubernetes Web UI$`, ps.iShouldNotBeAbleToAccessTheKubernetesWebUI)
	ctx.Step(`^the Kubernetes Web UI is disabled$`, ps.theKubernetesWebUIIsDisabled)

	ctx.AfterScenario(func(s *godog.Scenario, err error) {
		ps.tearDown()
		probes.LogScenarioEnd(s)
	})
}
