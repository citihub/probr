variable "prefix" {
  default     = "probr-demo"
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
  default     = ""
  description = "Filepath for kube config to be written to"
}
