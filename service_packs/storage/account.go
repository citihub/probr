package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2019-04-01/storage"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"

	"github.com/citihub/probr/internal/azureutil"
	"github.com/citihub/probr/internal/config"
)

//
func DeleteAccount(ctx context.Context, resourceGroupName, accountName string) error {

	c := accountClient()

	_, err := c.Delete(ctx, resourceGroupName, accountName)

	return err
}

// CreateWithNetworkRuleSet starts creation of a new Storage Account and waits for the account to be created.
func CreateWithNetworkRuleSet(ctx context.Context, accountName, accountGroupName string, tags map[string]*string, httpsOnly bool, networkRuleSet *storage.NetworkRuleSet) (storage.Account, error) {

	var sa storage.Account
	c := accountClient()

	r, err := c.CheckNameAvailability(
		ctx,
		storage.AccountCheckNameAvailabilityParameters{
			Name: to.StringPtr(accountName),
			Type: to.StringPtr("Microsoft.Storage/storageAccounts"),
		})
	if err != nil {
		return sa, err
	}

	if *r.NameAvailable != true {
		return sa, fmt.Errorf(
			"storage account name [%sa] not available: %v\nserver message: %v",
			accountName, err, *r.Message)
	}

	networkRuleSetParam := &storage.AccountPropertiesCreateParameters{
		EnableHTTPSTrafficOnly: to.BoolPtr(httpsOnly),
		NetworkRuleSet:         networkRuleSet,
	}

	future, err := c.Create(
		ctx,
		accountGroupName,
		accountName,
		storage.AccountCreateParameters{
			Sku: &storage.Sku{
				Name: storage.StandardLRS},
			Kind:                              storage.Storage,
			Location:                          to.StringPtr(azureutil.Location()),
			AccountPropertiesCreateParameters: networkRuleSetParam,
			Tags:                              tags,
		})

	if err != nil {
		return sa, err
	}

	err = future.WaitForCompletionRef(ctx, c.Client)
	if err != nil {
		return sa, err
	}

	return future.Result(c)
}

// AccountProperties returns the properties for the specified storage account including but not limited to name, SKU name, location, and account status
func AccountProperties(ctx context.Context, rgName, accountName string) (storage.Account, error) {
	return accountClient().GetProperties(ctx, rgName, accountName, "")
}

// AccountPrimaryKey return the primary key
func AccountPrimaryKey(ctx context.Context, accountName, accountGroupName string) string {
	response, err := getAccountKeys(ctx, accountName, accountGroupName)
	if err != nil {
		log.Fatalf("failed to list keys: %v", err)
	}
	return *(((*response.Keys)[0]).Value)
}

func getAccountKeys(ctx context.Context, accountName, accountGroupName string) (storage.AccountListKeysResult, error) {
	return accountClient().ListKeys(ctx, accountGroupName, accountName, "")
}

func accountClient() storage.AccountsClient {
	c := storage.NewAccountsClient(config.Vars.CloudProviders.Azure.SubscriptionID)

	// Check that connection config vars have been set
	if config.Vars.CloudProviders.Azure.TenantID == "" {
		log.Printf("[ERROR] Mandatory azure connection config var not set: config.Vars.CloudProviders.Azure.TenantID")
		return c
	}
	if config.Vars.CloudProviders.Azure.SubscriptionID == "" {
		log.Printf("[ERROR] Mandatory azure connection config var not set: config.Vars.CloudProviders.Azure.SubscriptionID")
		return c
	}
	if config.Vars.CloudProviders.Azure.ClientID == "" {
		log.Printf("[ERROR] Mandatory azure connection config var not set: config.Vars.CloudProviders.Azure.ClientID")
		return c
	}
	if config.Vars.CloudProviders.Azure.ClientSecret == "" {
		log.Printf("[ERROR] Mandatory azure connection config var not set: config.Vars.CloudProviders.Azure.ClientSecret")
		return c
	}

	authorizer := auth.NewClientCredentialsConfig(config.Vars.CloudProviders.Azure.ClientID, config.Vars.CloudProviders.Azure.ClientSecret, config.Vars.CloudProviders.Azure.TenantID)

	authorizerToken, err := authorizer.Authorizer()
	if err == nil {
		c.Authorizer = authorizerToken
	} else {
		log.Printf("[ERROR] Unable to authorise Storage Account accountClient: %v", err)
	}
	return c
}
