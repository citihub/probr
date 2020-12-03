package encryption_in_flight_azure

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	azurePolicy "github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-01-01/policy"
	azureStorage "github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2019-04-01/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"

	"github.com/citihub/probr/internal/azureutil"
	"github.com/citihub/probr/internal/azureutil/group"
	"github.com/citihub/probr/internal/azureutil/policy"
	"github.com/citihub/probr/internal/azureutil/storage"
)

const (
	policyName = "deny_http_storage"
)

// EncryptionInFlightAzure azure implementation of the encryption in flight for Object Storage feature
type EncryptionInFlightAzure struct {
	ctx                       context.Context
	tags                      map[string]*string
	httpOption                bool
	httpsOption               bool
	policyAssignmentMgmtGroup string
}

func (state *EncryptionInFlightAzure) setup() {
	log.Println("[DEBUG] Setting up \"EncryptionInFlightAzure\"")
	state.ctx = context.Background()
	state.policyAssignmentMgmtGroup = os.Getenv(azureutil.PolicyAssignmentManagementGroup)
	if state.policyAssignmentMgmtGroup == "" {
		log.Printf("[ERROR] '%v' environment variable is not defined. Policy assignment check against subscription", azureutil.PolicyAssignmentManagementGroup)
	}

	state.tags = map[string]*string{
		"project": to.StringPtr("CICD"),
		"env":     to.StringPtr("test"),
		"tier":    to.StringPtr("internal"),
	}

	_, err := group.CreateWithTags(state.ctx, azureutil.ResourceGroup(), state.tags)

	if err != nil {
		log.Fatalf("failed to create group: %v\n", err.Error())
	}
	log.Printf("[DEBUG] Created Resource Group: '%v'", azureutil.ResourceGroup())

}

func (state *EncryptionInFlightAzure) teardown() {
	group.Cleanup(state.ctx)
	log.Println("[DEBUG] Teardown completed")
}

func (state *EncryptionInFlightAzure) securityControlsThatRestrictDataFromBeingUnencryptedInFlight() error {
	var policyAssignment azurePolicy.Assignment
	var aerr error
	// Search assignment from Management Group instead of subscription
	if state.policyAssignmentMgmtGroup != "" {
		policyAssignment, aerr = policy.AssignmentByManagementGroup(state.ctx, state.policyAssignmentMgmtGroup, policyName)
	} else {
		policyAssignment, aerr = policy.AssignmentBySubscription(state.ctx, azureutil.SubscriptionID(), policyName)
	}

	if aerr != nil {
		log.Printf("[ERROR] Get policy assignment error: %v", aerr)
		return aerr
	}

	log.Printf("[DEBUG] Policy assignment check: %v [Step PASSED]", *policyAssignment.Name)
	return nil
}

func (state *EncryptionInFlightAzure) weProvisionAnObjectStorageBucket() error {
	// Nothing to do here
	return nil
}

func (state *EncryptionInFlightAzure) httpAccessIs(arg1 string) error {
	if arg1 == "enabled" {
		state.httpOption = true
	} else {
		state.httpOption = false
	}
	return nil
}

func (state *EncryptionInFlightAzure) httpsAccessIs(arg1 string) error {
	if arg1 == "enabled" {
		state.httpsOption = true
	} else {
		state.httpsOption = false
	}
	return nil
}

func (state *EncryptionInFlightAzure) creationWillWithAnErrorMatching(expectation, errDescription string) error {
	accountName := azureutil.RandString(5) + "storageac"

	var err error

	networkRuleSet := azureStorage.NetworkRuleSet{
		DefaultAction: azureStorage.DefaultActionDeny,
		IPRules:       &[]azureStorage.IPRule{},
	}

	// Both true take it as http option is try
	if state.httpsOption && state.httpOption {
		log.Printf("[DEBUG] Creating Storage Account with HTTPS: %v", false)
		_, err = storage.CreateWithNetworkRuleSet(state.ctx, accountName,
			azureutil.ResourceGroup(), state.tags, false, &networkRuleSet)
	} else if state.httpsOption {
		log.Printf("[DEBUG] Creating Storage Account with HTTPS: %v", state.httpsOption)
		_, err = storage.CreateWithNetworkRuleSet(state.ctx, accountName,
			azureutil.ResourceGroup(), state.tags, state.httpsOption, &networkRuleSet)
	} else if state.httpOption {
		log.Printf("[DEBUG] Creating Storage Account with HTTPS: %v", state.httpsOption)
		_, err = storage.CreateWithNetworkRuleSet(state.ctx, accountName,
			azureutil.ResourceGroup(), state.tags, state.httpsOption, &networkRuleSet)
	}

	if expectation == "Fail" {

		if err == nil {
			return fmt.Errorf("storage account was created, but should not have been: policy is not working or incorrectly configured")
		}

		detailedError := err.(autorest.DetailedError)
		originalErr := detailedError.Original
		detailed := originalErr.(*azure.ServiceError)

		log.Printf("[DEBUG] Detailed Error: %v", detailed)

		if strings.EqualFold(detailed.Code, "RequestDisallowedByPolicy") {
			// Now check if it is the right policy
			if strings.Contains(detailed.Message, policyName) {
				log.Printf("[DEBUG] Request was Disallowed By Policy: %v [Step PASSED]", policyName)
				return nil
			}
			return fmt.Errorf("storage account was not created but blocked not by the right policy: %v", detailed.Message)
		}

		return fmt.Errorf("storage account was not created")
	} else if expectation == "Succeed" {
		if err != nil {
			log.Printf("[ERROR] Unexpected failure in create storage ac [Step FAILED]")
			return err
		}
		return nil
	}

	return fmt.Errorf("unsupported `result` option '%s' in the Gherkin feature - use either 'Fail' or 'Succeed'", expectation)
}

func (state *EncryptionInFlightAzure) detectObjectStorageUnencryptedTransferAvailable() error {
	return fmt.Errorf("azure policy prevent creation of object storage with in-secure transport making this test irrelevant")
}

func (state *EncryptionInFlightAzure) detectObjectStorageUnencryptedTransferEnabled() error {
	return nil
}

func (state *EncryptionInFlightAzure) createUnencryptedTransferObjectStorage() error {
	return nil
}

func (state *EncryptionInFlightAzure) detectsTheObjectStorage() error {
	return nil
}

func (state *EncryptionInFlightAzure) encryptedDataTrafficIsEnforced() error {
	return nil
}
