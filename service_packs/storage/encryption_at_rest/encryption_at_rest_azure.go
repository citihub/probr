package encryption_at_rest_azure

import (
	"log"
)

// EncryptionAtRestAzure Azure implementation of the encryption in flight for Object Storage feature
type EncryptionAtRestAzure struct {
}

func (state *EncryptionAtRestAzure) securityControlsThatRestrictDataFromBeingUnencryptedAtRest() error {
	// It is available
	log.Printf("[DEBUG] Azure Storage account is encrypted by default and cannot be turned off. No test to run. Checking Azure Policy. (Unless customise this test to check for specific key usage.")
	return nil
}

func (state *EncryptionAtRestAzure) weProvisionAnObjectStorageBucket() error {
	return nil
}
func (state *EncryptionAtRestAzure) encryptionAtRestIs(encryptionOption string) error {
	return nil
}
func (state *EncryptionAtRestAzure) creationWillWithAnErrorMatching(result string) error {
	return nil
}

func (state *EncryptionAtRestAzure) createContainerWithoutEncryption() error {
	return nil
}
func (state *EncryptionAtRestAzure) detectiveDetectsNonCompliant() error {
	return nil
}
func (state *EncryptionAtRestAzure) containerIsRemediated() error {
	return nil
}

func (state *EncryptionAtRestAzure) setup() {
}

func (state *EncryptionAtRestAzure) teardown() {
}

func (state *EncryptionAtRestAzure) policyOrRuleAvailable() error {
	// It is available
	log.Printf("[DEBUG] Azure Storage account is encrypted by default and cannot be turned off. No test to run. Checking Azure Policy. (Unless customise this test to check for specific key usage.")
	return nil
}

func (state *EncryptionAtRestAzure) checkPolicyOrRuleAssignment() error {
	return nil
}

func (state *EncryptionAtRestAzure) policyOrRuleAssigned() error {
	return nil
}

func (state *EncryptionAtRestAzure) prepareToCreateContainer() error {
	return nil
}

func (state *EncryptionAtRestAzure) createContainerWithEncryptionOption(encryptionOption string) error {
	return nil
}

func (state *EncryptionAtRestAzure) createResult(result string) error {
	return nil
}
