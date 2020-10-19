// Package general provides the implementation required to execute the feature-based test cases
// described in the the 'events' directory.
package probes

import (
	"fmt"
	"strings"

	"github.com/citihub/probr/internal/clouddriver/kubernetes"
	"github.com/citihub/probr/internal/config"
	"github.com/citihub/probr/internal/coreengine"
	"github.com/citihub/probr/internal/utils"
	"github.com/cucumber/godog"
)

const gen_name = "general"

// init() registers the feature tests descibed in this package with the test runner (coreengine.TestRunner) via the call
// to coreengine.AddTestHandler.  This links the test - described by the TestDescriptor - with the handler to invoke.  In
// this case, the general test handler is being used (probes.GodogTestHandler) and the GodogTest data provides the data
// require to execute the test.  Specifically, the data includes the Test Suite and Scenario Initializers from this package
// which will be called from probes.GodogTestHandler.  Note: a blank import at probr library level should be done to
// invoke this function automatically on initial load.
func init() {
	td := coreengine.TestDescriptor{Group: coreengine.Kubernetes,
		Category: coreengine.General, Name: gen_name}

	coreengine.AddTestHandler(td, &coreengine.GoDogTestTuple{
		Handler: GodogTestHandler,
		Data: &coreengine.GodogTest{
			TestDescriptor:       &td,
			TestSuiteInitializer: genTestSuiteInitialize,
			ScenarioInitializer:  genScenarioInitialize,
		},
	})
}

// BUG - This step doesn't run
//@CIS-5.1.3
func (p *probeState) iInspectTheThatAreConfigured(roleLevel string) error {
	var err error
	if roleLevel == "Cluster Roles" {
		l, e := kubernetes.GetKubeInstance().GetClusterRolesByResource("*")
		err = e
		p.wildcardRoles = l
	} else if roleLevel == "Roles" {
		l, e := kubernetes.GetKubeInstance().GetRolesByResource("*")
		err = e
		p.wildcardRoles = l
	}
	if err != nil {
		err = LogAndReturnError("error raised when retrieving '%v': %v", roleLevel, err)
	}

	description := fmt.Sprintf("Ensures that %s are configured. Retains wildcard roles in state for following steps. Passes if retrieval command does not have error.", roleLevel)
	p.audit.AuditProbeStep(p.name, description, p, err)
	return err
}

func (p *probeState) iShouldOnlyFindWildcardsInKnownAndAuthorisedConfigurations() error {
	//we strip out system/known entries in the cluster roles & roles call
	var err error
	wildcardCount := len(p.wildcardRoles.([]interface{}))
	if wildcardCount > 0 {
		err = LogAndReturnError("roles exist with wildcarded resources")
	}

	description := "Examines probe state's wildcard roles. Passes if no wildcard roles are found."
	p.audit.AuditProbeStep(p.name, description, p, err)

	return err
}

//@CIS-5.6.3
func (p *probeState) iAttemptToCreateADeploymentWhichDoesNotHaveASecurityContext() error {
	cname := "probr-general"
	pod_name := kubernetes.GenerateUniquePodName(cname)
	image := config.Vars.Images.Repository + "/" + config.Vars.Images.BusyBox

	//create pod with nil security context
	pod, podAudit, err := kubernetes.GetKubeInstance().CreatePod(pod_name, "probr-general-test-ns", cname, image, true, nil)

	err = ProcessPodCreationResult(p.name, &p.state, pod, kubernetes.UndefinedPodCreationErrorReason, err)

	description := "Attempts to create a deployment without a security context. Retains the status of the deployment in probe state for following steps. Passes if created, or if an expected error is encountered."

	payload := podPayload(pod, podAudit)
	p.audit.AuditProbeStep(p.name, description, payload, err)
	return err
}

func (p *probeState) theDeploymentIsRejected() error {
	//looking for a non-nil creation error
	var err error
	if p.state.CreationError == nil {
		err = LogAndReturnError("pod %v was created successfully. Test fail.", p.state.PodName)
	}

	description := "Looks for a creation error on the current probe state. Passes if error is found, because it should have been rejected."
	p.audit.AuditProbeStep(p.name, description, nil, err)

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
		err = LogAndReturnError("error raised when trying to retrieve pods: %v", err)
	} else {
		//a "pass" is the abscence of a "kubernetes-dashboard" pod
		for _, v := range pl.Items {
			if strings.HasPrefix(v.Name, "kubernetes-dashboard") {
				err = LogAndReturnError("kubernetes-dashboard pod found (%v) - test fail", v.Name)
			}
		}
	}

	description := "Attempts to find a pod in the 'kube-system' namespace with the prefix 'kubernetes-dashboard'. Passes if no pod is found."
	p.audit.AuditProbeStep(p.name, description, nil, err)

	return err
}

// genTestSuiteInitialize handles any overall Test Suite initialisation steps.  This is registered with the
// test handler as part of the init() function.
func genTestSuiteInitialize(ctx *godog.TestSuiteContext) {

	ctx.BeforeSuite(func() {}) //nothing for now

	ctx.AfterSuite(func() {})

}

// genScenarioInitialize initialises the specific test steps.  This is essentially the creation of the test
// which reflects the tests described in the events directory.  There must be a test step registered for
// each line in the feature files. Note: Godog will output stub steps and implementations if it doesn't find
// a step / function defined.  See: https://github.com/cucumber/godog#example.
func genScenarioInitialize(ctx *godog.ScenarioContext) {
	ps := probeState{}

	ctx.BeforeScenario(func(s *godog.Scenario) {
		ps.BeforeScenario(gen_name, s)
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
		kubernetes.GetKubeInstance().DeletePod(&ps.state.PodName, utils.StringPtr("probr-general-test-ns"), false, gen_name)
		LogScenarioEnd(s)
	})
}
