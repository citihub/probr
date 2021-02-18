@k-gen
@probes/kubernetes/general
Feature: General Cluster Security Configurations
  As a Security Auditor
  I want to ensure that Kubernetes clusters have general security configurations in place
  So that no general cluster vulnerabilities can be exploited

    @k-gen-001
    Scenario: Ensure Kubernetes Web UI is disabled

    The Kubernetes Web UI (Dashboard) has been a historical source of vulnerability and should only be deployed when necessary.

      Given a Kubernetes cluster is deployed
      Then the Kubernetes Web UI is disabled
