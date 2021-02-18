@k-iam
@probes/kubernetes/iam
Feature: Ensure stringent authentication and authorisation
    As a Security Auditor
    I want to ensure that stringent authentication and authorisation policies are applied to my organisation's Kubernetes clusters
    So that only approved actors have the ability to perform sensitive operations in order to prevent malicious attacks on my organization

    Background:
        Given a Kubernetes cluster is deployed
        And the cluster has managed identity components deployed

    @k-iam-001
    Scenario Outline: Prevent cross namespace Azure Identities
        # TODO PJITREVIEW remove implementation detail from here
        And an AzureIdentityBinding called "probr-aib" exists in the namespace called "default"
        When I create a simple pod in "<NAMESPACE>" namespace assigned with the "probr-aib" AzureIdentityBinding
        Then the pod is deployed successfully
        But an attempt to obtain an access token from that pod should "<RESULT>"

        Examples:
            | NAMESPACE   | RESULT  |
            | the probr   | Fail    |
            | the default | Succeed |

    @k-iam-002
    Scenario: Prevent cross namespace Azure Identity Bindings
        # TODO PJITREVIEW remove implementation detail from here
        # TODO PJITREVIEW add additional Given to satisfy the positive case
        # Use more general terms like "my namespace" "another namespace"
        And the namespace called "default" has an AzureIdentity called "probr-probe"
        When I create an AzureIdentityBinding called "probr-aib" in the Probr namespace bound to the "probr-probe" AzureIdentity
        And I deploy a pod assigned with the "probr-aib" AzureIdentityBinding into the Probr namespace
        Then the pod is deployed successfully
        But an attempt to obtain an access token from that pod should fail

    @k-iam-003
    # TODO PJITREVIEW needs cluster reader or at least reader on the MIC namespace
    # e.g. @permissions/clusterreaderrole
    Scenario: Prevent access to AKS credentials via Azure Identity Components
    
    On the agent node VMs in the Kubernetes cluster, service principal credentials are stored in the file
    /etc/kubernetes/azure.json, which should therefore be protected.

    See https://docs.microsoft.com/en-us/azure/aks/kubernetes-service-principal

        When I execute the command "cat /etc/kubernetes/azure.json" against the MIC pod
        Then Kubernetes should prevent me from running the command


    @k-iam-004
    Scenario Outline: Minimise wildcards in Roles and Cluster Roles

    Kubernetes roles provide access to resources. Using wildcards does not adhere to the security principle of least
    privilege. Other than known system-assigned and well understood role definitions (which are configurable in Probr),
    we should not expect to use wildcards.

        When I inspect the "<rolelevel>" that are configured
        Then I should only find wildcards in known and authorised configurations

        Examples:
            | rolelevel     |
            | Roles         |
            | Cluster Roles |