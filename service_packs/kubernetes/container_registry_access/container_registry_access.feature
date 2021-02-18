@k-cra
@probes/kubernetes/container_registry_access
Feature: Protect image container registries
  As a Security Auditor
  I want to ensure that containers image registries are secured in my organisation's Kubernetes clusters
  So that only approved software can be run in our cluster in order to prevent malicious attacks on my organization

    #Rule: CHC2-APPDEV135 - Ensure software release and deployment is managed through a formal, controlled process

    Background:
      Given a Kubernetes cluster is deployed

    @k-cra-001
    Scenario: Ensure the cluster service account has read only access to the authorized container registry
      When I attempt to push to the container registry using the cluster identity
      Then the push request is rejected due to authorization

    @k-cra-002
    Scenario: Ensure deployment from an unauthorised container registry is denied
      Given a user attempts to deploy a container from an authorised registry
      And the deployment attempt is allowed
      When a user attempts to deploy a container from an unauthorised registry
      Then the deployment attempt is denied
