package encryption_at_rest

import (
	"context"
	"log"

	"github.com/citihub/probr/internal/coreengine"
	"github.com/citihub/probr/internal/summary"
	"github.com/cucumber/godog"
)

type scenarioState struct {
	name                      string
	audit                     *summary.ScenarioAudit
	probe                     *summary.Probe
	ctx                       context.Context
	tags                      map[string]*string
	httpOption                bool
	httpsOption               bool
	policyAssignmentMgmtGroup string
}

// Allows this probe to be added to the ProbeStore
type ProbeStruct struct {
	state scenarioState
}

// Allows this probe to be added to the ProbeStore
var Probe ProbeStruct

func (state *scenarioState) securityControlsThatRestrictDataFromBeingUnencryptedAtRest() error {
	// It is available
	log.Printf("[DEBUG] Azure Storage account is encrypted by default and cannot be turned off. No test to run. Checking Azure Policy. (Unless customise this test to check for specific key usage.")
	return nil
}

// PENDING IMPLEMENTATION
func (state *scenarioState) weProvisionAnObjectStorageBucket() error {
	return nil
}

// PENDING IMPLEMENTATION
func (state *scenarioState) encryptionAtRestIs(encryptionOption string) error {
	return nil
}

// PENDING IMPLEMENTATION
func (state *scenarioState) creationWillWithAnErrorMatching(result string) error {
	return nil
}

// PENDING IMPLEMENTATION
func (state *scenarioState) createContainerWithoutEncryption() error {
	return nil
}

// PENDING IMPLEMENTATION
func (state *scenarioState) detectiveDetectsNonCompliant() error {
	return nil
}

// PENDING IMPLEMENTATION
func (state *scenarioState) containerIsRemediated() error {
	return nil
}

// PENDING IMPLEMENTATION
func (state *scenarioState) setup() {
}

// PENDING IMPLEMENTATION
func (state *scenarioState) teardown() {
}

// PENDING IMPLEMENTATION
func (state *scenarioState) policyOrRuleAvailable() error {
	// It is available
	log.Printf("[DEBUG] Azure Storage account is encrypted by default and cannot be turned off. No test to run. Checking Azure Policy. (Unless customise this test to check for specific key usage.")
	return nil
}

// PENDING IMPLEMENTATION
func (state *scenarioState) checkPolicyOrRuleAssignment() error {
	return nil
}

// PENDING IMPLEMENTATION
func (state *scenarioState) policyOrRuleAssigned() error {
	return nil
}

// PENDING IMPLEMENTATION
func (state *scenarioState) prepareToCreateContainer() error {
	return nil
}

// PENDING IMPLEMENTATION
func (state *scenarioState) createContainerWithEncryptionOption(encryptionOption string) error {
	return nil
}

// PENDING IMPLEMENTATION
func (state *scenarioState) createResult(result string) error {
	return nil
}

func (s *scenarioState) beforeScenario(probeName string, gs *godog.Scenario) {
	s.name = gs.Name
	s.probe = summary.State.GetProbeLog(probeName)
	s.audit = summary.State.GetProbeLog(probeName).InitializeAuditor(gs.Name, gs.Tags)
	coreengine.LogScenarioStart(gs)
}

// Return this probe's name
func (p ProbeStruct) Name() string {
	return "encryption_at_rest"
}

func (p ProbeStruct) Path() string {
	return coreengine.GetFeaturePath("service_packs", "storage", "azure", p.Name())
}

// ProbeInitialize handles any overall Test Suite initialisation steps.  This is registered with the
// test handler as part of the init() function.
//func (p ProbeStruct) ProbeInitialize(ctx *godog.Suite) {
func (p ProbeStruct) ProbeInitialize(ctx *godog.TestSuiteContext) {
	p.state = scenarioState{}

	ctx.BeforeSuite(p.state.setup)

	ctx.AfterSuite(p.state.teardown)
}

// initialises the scenario
func (p ProbeStruct) ScenarioInitialize(ctx *godog.ScenarioContext) {

	ctx.BeforeScenario(func(s *godog.Scenario) {
		p.state.beforeScenario(p.Name(), s)
	})

	ctx.Step(`^security controls that restrict data from being unencrypted at rest$`, p.state.securityControlsThatRestrictDataFromBeingUnencryptedAtRest)
	ctx.Step(`^we provision an Object Storage bucket$`, p.state.weProvisionAnObjectStorageBucket)
	ctx.Step(`^encryption at rest is "([^"]*)"$`, p.state.encryptionAtRestIs)
	ctx.Step(`^creation will "([^"]*)" with an error matching "([^"]*)"$`, p.state.creationWillWithAnErrorMatching)

	ctx.Step(`^there is a detective capability for creation of Object Storage without encryption at rest$`, p.state.policyOrRuleAvailable)
	ctx.Step(`^the capability for detecting the creation of Object Storage without encryption at rest is active$`, p.state.checkPolicyOrRuleAssignment)
	ctx.Step(`^the detective measure is enabled$`, p.state.policyOrRuleAssigned)
	ctx.Step(`^Object Storage is created with without encryption at rest$`, p.state.createContainerWithoutEncryption)
	ctx.Step(`^the detective capability detects the creation of Object Storage without encryption at rest$`, p.state.detectiveDetectsNonCompliant)
	ctx.Step(`^the detective capability enforces encryption at rest on the Object Storage Bucket$`, p.state.containerIsRemediated)

	ctx.AfterScenario(func(s *godog.Scenario, err error) {
		coreengine.LogScenarioEnd(s)
	})
}
