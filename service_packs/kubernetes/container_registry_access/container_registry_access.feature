@k-cra
@probes/kubernetes/container_registry_access
Feature: Protect image container registries
    As a Security Auditor
    I want to ensure that containers image registries are secured in my organisation's Kubernetes clusters
    So that only approved software can be run in our cluster in order to prevent malicious attacks on my organization

    #Rule: CHC2-APPDEV135 - Ensure software release and deployment is managed through a formal, controlled process

    Background:
        Given a Kubernetes cluster exists which we can deploy into

    @k-cra-002
    Scenario: Ensure deployment from an authorised container registry is allowed
        Then pod creation "succeeds" with container image from "authorized" registry
        And pod creation "is denied" with container image from "unauthorized" registry

    # @k-cra-002
    # Scenario: Ensure deployment from an authorised container registry is allowed
        #When a user attempts to deploy a container from an authorised registry
        #Then the deployment attempt is allowed
        
    # @k-cra-003
    # Scenario: Ensure deployment from an unauthorised container registry is denied
    #   When a user attempts to deploy a container from an unauthorised registry
    #   Then the deployment attempt is denied
