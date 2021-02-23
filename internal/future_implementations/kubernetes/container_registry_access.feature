    @k-cra-001
    Scenario: Ensure container image registries are read-only
        Given a Kubernetes cluster is deployed
        And I am authorised to pull from a container registry
        When I attempt to push to the container registry using the cluster identity
        Then the push request is rejected due to authorization
