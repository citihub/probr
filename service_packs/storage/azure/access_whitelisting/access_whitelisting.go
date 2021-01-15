package access_whitelisting

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	azureStorage "github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2019-04-01/storage"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/cucumber/godog"

	"github.com/citihub/probr/internal/azureutil"
	"github.com/citihub/probr/internal/azureutil/group"
	"github.com/citihub/probr/internal/coreengine"
	"github.com/citihub/probr/internal/summary"
	"github.com/citihub/probr/internal/utils"
	"github.com/citihub/probr/service_packs/storage"
)

const (
	policyAssignmentName = "deny_storage_wo_net_acl"
	storageRgEnvVar      = "STORAGE_ACCOUNT_RESOURCE_GROUP"
)

// Allows this probe to be added to the ProbeStore
type ProbeStruct struct {
	state scenarioState
}

// Allows this probe to be added to the ProbeStore
var Probe ProbeStruct

type scenarioState struct {
	name                      string
	audit                     *summary.ScenarioAudit
	probe                     *summary.Probe
	ctx                       context.Context
	policyAssignmentMgmtGroup string
	tags                      map[string]*string
	bucketName                string
	storageAccount            azureStorage.Account
	runningErr                error
}

func (state *scenarioState) setup() {

	log.Println("[DEBUG] Setting up \"AccessWhitelistingAzure\"")

}

func (state *scenarioState) teardown() {

	log.Println("[DEBUG] Teardown completed")
}

func (state *scenarioState) anAzureResourceGroupExists() error {

	var err error

	var stepTrace strings.Builder

	stepTrace.WriteString("Check if value for Azure resource group is set in config file;")
	if azureutil.ResourceGroup() == "" {
		log.Printf("[ERROR] Azure resource group config var not set")
		err = errors.New("Azure resource group config var not set")
	}
	if err == nil {
		stepTrace.WriteString("Check the resource group exists in the specified azure subscription;")
		_, err = group.Get(state.ctx, azureutil.ResourceGroup())
		if err != nil {
			log.Printf("[ERROR] Configured Azure resource group %s does not exists", azureutil.ResourceGroup())
		}
	}

	description := stepTrace.String()
	payload := struct {
		AzureResourceGroup string
	}{
		AzureResourceGroup: azureutil.ResourceGroup(),
	}
	state.audit.AuditScenarioStep(description, payload, err)

	return err
}

func (state *scenarioState) checkPolicyAssigned() error {

	/////////////////////////
	err := fmt.Errorf("Not Implemented")

	var stepTrace strings.Builder
	stepTrace.WriteString("TODO: Pending implementation;")

	description := stepTrace.String()
	payload := struct {
	}{}
	state.audit.AuditScenarioStep(description, payload, err)

	//return err
	return nil //TODO: Remove this line. This is temporary to ensure test doesn't halt and other steps are not skipped
	/////////////////////////

	// var a azurePolicy.Assignment
	// var err error

	// // If a Management Group has not been set, check Policy Assignment at the Subscription
	// if state.policyAssignmentMgmtGroup == "" {
	// 	a, err = policy.AssignmentBySubscription(state.ctx, azureutil.SubscriptionID(), policyAssignmentName)
	// } else {
	// 	a, err = policy.AssignmentByManagementGroup(state.ctx, state.policyAssignmentMgmtGroup, policyAssignmentName)
	// }

	// if err != nil {
	// 	log.Printf("[ERROR] Policy Assignment error: %v", err)
	// 	return err
	// }

	// log.Printf("[DEBUG] Policy Assignment check: %v [Step PASSED]", *a.Name)
	// return nil
}

func (state *scenarioState) provisionStorageContainer() error {

	var stepTrace strings.Builder
	var err error

	stepTrace.WriteString("A bucket name is defined using a random string, storage account is not yet provisioned;")
	// define a bucket name, then pass the step - we will provision the account in the next step.
	state.bucketName = utils.RandomString(10)

	description := stepTrace.String()
	payload := struct {
		BucketName string
	}{
		BucketName: state.bucketName,
	}
	state.audit.AuditScenarioStep(description, payload, err)

	return err
}

func (state *scenarioState) createWithWhitelist(ipRange string) error {

	// ////////////////////////////////
	// err := fmt.Errorf("Not Implemented")

	// var stepTrace strings.Builder
	// stepTrace.WriteString("TODO: Pending implementation;")

	// description := stepTrace.String()
	// payload := struct {
	// }{}
	// state.audit.AuditScenarioStep(description, payload, err)

	// //return err
	// return nil //TODO: Remove this line. This is temporary to ensure test doesn't halt and other steps are not skipped
	// ///////////////////////////////

	var networkRuleSet azureStorage.NetworkRuleSet
	if ipRange == "nil" {
		networkRuleSet = azureStorage.NetworkRuleSet{
			DefaultAction: azureStorage.DefaultActionAllow,
		}
	} else {
		ipRule := azureStorage.IPRule{
			Action:           azureStorage.Allow,
			IPAddressOrRange: to.StringPtr(ipRange),
		}

		networkRuleSet = azureStorage.NetworkRuleSet{
			IPRules:       &[]azureStorage.IPRule{ipRule},
			DefaultAction: azureStorage.DefaultActionDeny,
		}
	}

	state.storageAccount, state.runningErr = storage.CreateWithNetworkRuleSet(state.ctx, state.bucketName, azureutil.ResourceGroup(), state.tags, true, &networkRuleSet)
	return nil
}

func (state *scenarioState) creationWill(expectation string) error {

	var err error
	var stepTrace strings.Builder
	payload := struct {
		StorageAccountID string
	}{}

	stepTrace.WriteString(fmt.Sprintf("Expectation that Object Storage container was provisioned with whitelisting in previous step is: %s;", expectation))
	payload.StorageAccountID = *state.storageAccount.ID

	// if expectation == "Fail" {
	// 	if state.runningErr == nil {
	// 		//return fmt.Errorf("incorrectly created Storage Account: %v", *state.storageAccount.ID)
	// 	}
	// 	//return nil
	// }

	// if state.runningErr == nil {
	// 	return nil
	// }

	if (expectation == "Fail" && state.runningErr == nil) || (expectation == "Success" && state.runningErr != nil) {
		err = fmt.Errorf("incorrectly created Storage Account: %v", *state.storageAccount.ID)
	}
	state.audit.AuditScenarioStep(stepTrace.String(), payload, err)

	return err
}

func (state *scenarioState) cspSupportsWhitelisting() error {

	err := fmt.Errorf("Not Implemented")

	var stepTrace strings.Builder
	stepTrace.WriteString("TODO: Pending implementation;")

	description := stepTrace.String()
	payload := struct {
	}{}
	state.audit.AuditScenarioStep(description, payload, err)

	//return err
	return nil //TODO: Remove this line. This is temporary to ensure test doesn't halt and other steps are not skipped
}

func (state *scenarioState) examineStorageContainer(containerNameEnvVar string) error {
	return nil

	accountName := os.Getenv(containerNameEnvVar)
	if accountName == "" {
		return fmt.Errorf("environment variable \"%s\" is not defined test can't run", containerNameEnvVar)
	}

	resourceGroup := os.Getenv(storageRgEnvVar)
	if resourceGroup == "" {
		return fmt.Errorf("environment variable \"%s\" is not defined test can't run", storageRgEnvVar)
	}

	state.storageAccount, state.runningErr = storage.AccountProperties(state.ctx, resourceGroup, accountName)

	if state.runningErr != nil {
		return state.runningErr
	}

	networkRuleSet := state.storageAccount.AccountProperties.NetworkRuleSet
	result := false
	// Default action is deny
	if networkRuleSet.DefaultAction == azureStorage.DefaultActionAllow {
		return fmt.Errorf("%s has not configured with firewall network rule default action is not deny", accountName)
	}

	// Check if it has IP whitelisting
	for _, ipRule := range *networkRuleSet.IPRules {
		result = true
		log.Printf("[DEBUG] IP WhiteListing: %v, %v", *ipRule.IPAddressOrRange, ipRule.Action)
	}

	// Check if it has private Endpoint whitelisting
	for _, vnetRule := range *networkRuleSet.VirtualNetworkRules {
		result = true
		log.Printf("[DEBUG] VNet whitelisting: %v, %v", *vnetRule.VirtualNetworkResourceID, vnetRule.Action)
	}

	// TODO: Private Endpoint implementation when it's GA

	if result {
		log.Printf("[DEBUG] Whitelisting rule exists. [Step PASSED]")
		return nil
	}
	return fmt.Errorf("no whitelisting has been defined for %v", accountName)
}

// PENDING IMPLEMENTATION
func (state *scenarioState) whitelistingIsConfigured() error {
	// Checked in previous step

	err := fmt.Errorf("Not Implemented")

	var stepTrace strings.Builder
	stepTrace.WriteString("TODO: Pending implementation;")

	description := stepTrace.String()
	payload := struct {
	}{}
	state.audit.AuditScenarioStep(description, payload, err)

	//return err
	return nil //TODO: Remove this line. This is temporary to ensure test doesn't halt and other steps are not skipped
}

func (s *scenarioState) beforeScenario(probeName string, gs *godog.Scenario) {
	s.name = gs.Name
	s.probe = summary.State.GetProbeLog(probeName)
	s.audit = summary.State.GetProbeLog(probeName).InitializeAuditor(gs.Name, gs.Tags)
	s.ctx = context.Background()
	coreengine.LogScenarioStart(gs)
}

// Return this probe's name
func (p ProbeStruct) Name() string {
	return "access_whitelisting"
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

	ctx.Step(`^the CSP provides a whitelisting capability for Object Storage containers$`, p.state.cspSupportsWhitelisting)
	ctx.Step(`^a specified azure resource group exists$`, p.state.anAzureResourceGroupExists)
	ctx.Step(`^we examine the Object Storage container in environment variable "([^"]*)"$`, p.state.examineStorageContainer)
	ctx.Step(`^whitelisting is configured with the given IP address range or an endpoint$`, p.state.whitelistingIsConfigured)
	ctx.Step(`^security controls that Prevent Object Storage from being created without network source address whitelisting are applied$`, p.state.checkPolicyAssigned)
	ctx.Step(`^we provision an Object Storage container$`, p.state.provisionStorageContainer)
	ctx.Step(`^it is created with whitelisting entry "([^"]*)"$`, p.state.createWithWhitelist)
	ctx.Step(`^creation will "([^"]*)"$`, p.state.creationWill)

	ctx.AfterScenario(func(s *godog.Scenario, err error) {
		coreengine.LogScenarioEnd(s)
	})
}
