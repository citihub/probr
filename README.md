# Probr

## Your Zero Trust Compliance Toolbox

Probr automatically interprets your cloud platform control requirements (specified in cucumber feature files) and 'probes' your target platform to test whether the controls have been correctly implemented.

In addition to vaildating that specific policies are in place, Probr performs actions in order to test that controls are effective. For example, an attempt is made to pull an image from an unauthorized container registry in order to validate that the action is correctly denied.

Probr is of use for **security professionals** and **engineering teams** to validate that internal policy and external regulatory requirements are being complied with. Typically, Probr will be run within the CI/CD pipline to ensure that any changes to the platform are automatically tested for compliance.

## Installation

1. Download the latest Probr package by clicking the corresponding asset on our [release page](https://github.com/citihub/probr/releases). This includes the source code and executable.
2. If necessary, build the edge version of Probr by using `go build -o probr.exe cmd/main.go` from the source code. This will be necessary if an executable compatible with your system is not available on the release page.
3. Install the Probr executable in your chosen working directory, and configure it by setting configuration variables as described in the `Configuration` section below.

## Usage

Run the probr executable via `./probr [options]`.

Command line options can be obtained via `./probr --help`

## Configuration

### How the Config Works

Configuration variables can be populated in one of four ways, with the value being taken from the highest priority entry.

1. Default values; found in `internal/config/defaults.go` (lowest priority)
1. OS environment variables; set locally prior to probr execution (mid priority)
1. Vars file; yaml (highest non-CLI priority)
1. CLI flags; see `./probr --help` for available flags (highest priority)

_Note: See `internal/config/README.md` for engineering notes regarding configuration._

### Environment Variables

If you would like to handle logic differently per environment, env vars may be useful. An example of how to set an env var is as follows:

`export KUBE_CONFIG=./path/to/config`

### Vars File

An example Vars file is available at `probr/examples/config.yml`
You may have as many vars files as you wish in your codebase, which will enable you to maintain configurations for multiple environments in a single codebase.

The location of the vars file is passed as a CLI option e.g.

```
probr --varsFile=./config-dev.yml
```

**IMPORTANT:** Remember to encrypt your config file if it contains secrets.

### Probr Configuration Variables

These are general configuration variables.

| Variable | Description | CLI Option | Vars File | Env Var | Default |
|---|---|---|---|---|---|
|VarsFile|Config YAML File Path|yes|N/A|N/A|N/A|
|Silent|Disable visual runtime indicator|yes|no|N/A|true|
|OutputType|Determines output to file (IO) or terminal (INMEM)|yes|yes|PROBR_OUTPUT_TYPE|INMEM|
|OutputDir|Path to output dir if applicable|yes|yes|PROBR_CUCUMBER_DIR|cucumber_output|
|Tags|Feature tag inclusions and exclusions|yes|yes|PROBR_TAGS|""|
|AuditEnabled|Flag to switch on audit log|no|yes|PROBR_AUDIT_ENABLED|true|
|SummaryEnabled|Flag to switch on summary log|no|yes|PROBR_SUMMARY_ENABLED|true|
|AuditDir|Path to audit dir|no|yes|PROBR_AUDIT_DIR|audit_output|
|LogLevel|Set log verbosity level|no|yes|PROBR_LOG_LEVEL|ERROR|
|OverwriteHistoricalAudits|Flag to allow audit overwriting|no|yes|OVERWRITE_AUDITS|true|
|ContainerRegistry|Probe image container regsitry|no|yes|PROBR_CONTAINER_REGISTRY|docker.io|
|ProbeImage|Probe image name|no|probeImage|PROBR_PROBE_IMAGE|citihub/probr-probe|

### Service Pack Configuration Variables

Variables that are specific to a service pack. May be configured in the Vars file via embedded tags under ServicePacks.

| Variable | Description | CLI Flag | VarsFile | Env Var | Default |
|---|---|---|---|---|---|
|Kubernetes.KubeConfig|Path to kubernetes config|yes|yes|KUBE_CONFIG|N/A|
|Kubernetes.KubeContext|Kubernetes context|no|yes|KUBE_CONTEXT|""|
|Kubernetes.SystemClusterRoles|Cluster names|no|yes|N/A|{"system:", "aks", "cluster-admin", "policy-agent"}|

### Cloud Provider Configuration Variables

Variables that are specific to a cloud service provider and can be configured in the Vars file via embedded tags under CloudProviders.

| Variable | Description | CLI Flag | VarsFile | Env Var | Default |
|---|---|---|---|---|---|
|Azure.SubscriptionID|Azure subscription|no|yes|AZURE_SUBSCRIPTION_ID|""|
|Azure.ClientId|Azure client id|no|yes|AZURE_CLIENT_ID|""|
|Azure.ClientSecret|Azure client secret|no|yes|AZURE_CLIENT_SECRET|""|
|Azure.TenantID|Azure tenant id|no|yes|AZURE_TENANT_ID|""|
|Azure.LocationDefault|Azure location default|no|yes|AZURE_LOCATION_DEFAULT|""|
|Azure.AzureIdentity.DefaultNamespaceAI|Azure namespace|no|yes|DEFAULT_NS_AZURE_IDENTITY|probr-defaultns-ai|
|Azure.AzureIdentity.DefaultNamespaceAIB|Azure namespace|no|yes|DEFAULT_NS_AZURE_IDENTITY_BINDING|probr-defaultns-aib|

## Development & Contributing

Please see the [contributing docs](https://github.com/citihub/probr/blob/master/CONTRIBUTING.md) for information on how to develop and contribute to this repository as either a maintainer or open source contributor (the same rules apply for both).
