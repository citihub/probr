package group

import (
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-02-01/resources"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"

	"github.com/citihub/probr/internal/azureutil"
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

// Cleanup deletes the Resource Group created during testing (a test Resource Group name in the form 'test[a-z]{6}resourceGP').
func Cleanup(ctx context.Context) error {
	log.Println("[DEBUG] Deleting resources")
	_, err := client().Delete(ctx, azureutil.ResourceGroup())
	return err
}

func client() resources.GroupsClient {
	c := resources.NewGroupsClient(azureutil.SubscriptionID())
	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err == nil {
		c.Authorizer = authorizer
	} else {
		log.Fatalf("Unable to authorise Resource Group client: %v", err)
	}
	return c
}
