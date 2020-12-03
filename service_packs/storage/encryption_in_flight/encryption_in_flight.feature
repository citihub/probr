@intrusive_test
@service.object_storage
@encryption_in_flight
@CCO:CHC2-SVD001
@CCO:CHC2-AGP140
@csp.aws
@csp.azure
Feature: Object Storage Encryption in Flight

  As a Cloud Security Architect
  I want to ensure that suitable security controls are applied to Object Storage
  So that my organisation is not vulnerable to interception of data in transit

  Rule: CHC2-AGP140 - Ensure cryptographic controls are in place to protect the confidentiality and integrity of data in-transit, stored, generated and processed in the cloud

    @preventative
    Scenario Outline: Prevent Creation of Object Storage Without Encryption in Flight
      Given security controls that restrict data from being unencrypted in flight
      When we provision an Object Storage bucket
      And http access is "<HTTP Option>"
      And https access is "<HTTPS Option>"
      Then creation will "<Result>" with an error matching "<Error Description>"

      Examples:
        | HTTP Option | HTTPS Option | Result  | Error Description                                     |
        | enabled     | disabled     | Fail    | Storage Buckets must not be accessible via plain HTTP |
        | enabled     | enabled      | Fail    | Storage Buckets must not be accessible via plain HTTP |
        | disabled    | enabled      | Succeed |                                                       |

  @detective
  Scenario: Remediate Object Storage if Creation of Object Storage Without Encryption in Flight is Detected
    Given there is a detective capability for creation of Object Storage with unencrypted data transfer enabled
    And the capability for detecting the creation of Object Storage with unencrypted data transfer enabled is active
    When Object Storage is created with unencrypted data transfer enabled
    Then the detective capability detects the creation of Object Storage with unencrypted data transfer enabled
    And the detective capability enforces encrypted data transfer on the Object Storage Bucket