package group

import (
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-02-01/resources"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"

	"github.com/citihub/probr/internal/azureutil"
	"github.com/citihub/probr/internal/config"
)

// Create creates a new Resource Group in the default location (configured using the AZURE_LOCATION environment variable).
func Create(ctx context.Context, name string) (resources.Group, error) {
	log.Printf("[DEBUG] creating Resource Group '%s' in location: %v", name, azureutil.Location())
	return client().CreateOrUpdate(
		ctx,
		name,
		resources.Group{
			Location: to.StringPtr(azureutil.Location()),
		})
}

// Get an existing Resource Group by name
func Get(ctx context.Context, name string) (resources.Group, error) {
	log.Printf("[DEBUG] getting a Resource Group '%s'", name)
	return client().Get(ctx, name)
}

// CreateWithTags creates a new Resource Group in the default location (configured using the AZURE_LOCATION environment variable) and sets the supplied tags.
func CreateWithTags(ctx context.Context, name string, tags map[string]*string) (resources.Group, error) {
	log.Printf("[DEBUG] creating Resource Group '%s' on location: '%v'", name, azureutil.Location())
	return client().CreateOrUpdate(
		ctx,
		name,
		resources.Group{
			Location: to.StringPtr(azureutil.Location()),
			Tags:     tags,
		})
}

func client() resources.GroupsClient {

	c := resources.NewGroupsClient(config.Vars.CloudProviders.Azure.SubscriptionID)

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
		log.Printf("[ERROR] Unable to authorise Resource Group client: %v", err)
	}
	return c
}
