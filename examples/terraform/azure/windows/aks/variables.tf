variable "azure_subscription"{
  default = "47a81967-e656-4c1a-859c-d0f3f27c4899"
  description = "Azure subscription to use"
}

variable "prefix" {
  default     = "probr-automation"
  description = "Value that will be prepended to most others"
}

variable "location" {
  default     = "East US 2"
  description = "Location display name. az account list-locations --output table"
}

variable "cluster_name" {
  default     = "cluster"
  description = "K8s cluster. Should recieve prefix."
}

variable "kube_config_filepath" {
  default     = "~/go/src/my-module/"
  description = "Filepath for kube config to be written to"
}

variable "demo_acr" {
  default = "automation"
}

variable "acr_name" {
  default = "marioprobr"
}
variable "demo_acr_rg" {
  default = "probr-automation-rg"
}

variable "probr_probe_msi_name" {
  default = "probr-msi"
}

variable "psp_policy_name" {
  default = "probr demo psp policy"
  description = "Restricted PSP policy"
}

variable "restrict_registry_policy_name" {
  default = "probr demo psp policy"
  description = "Restrict container registry"
}

variable "namespaces" {
  type = list

  default = [
    "probr-container-access-test-ns",
    "probr-general-test-ns",
    "probr-network-access-test-ns",
    "probr-pod-security-test-ns",
    "probr-rbac-test-ns"
  ]

}
