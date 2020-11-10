provider "azurerm" {
  version = "~>2.5.0"
  features {} // azurerm will err if this is not included
}

resource "azurerm_resource_group" "rg" {
  name     = "${var.prefix}-rg"
  location = var.location
}

resource "azurerm_kubernetes_cluster" "cluster" {
  name                = "${var.prefix}-${var.cluster_name}"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
  dns_prefix          = "${var.prefix}-dns"
  tags = {
    aadpodidentity : "enabled",
    policies : "all",
    project : "probr demo"
  }

  default_node_pool {
    name       = "default"
    node_count = 1
    vm_size    = "Standard_DS2_v2"
    //vnet_subnet_id = element(tolist(azurerm_virtual_network.vnet.subnet), 0).id // subnet object contains one value
  }

  identity {
    type = "SystemAssigned"
  }

  network_profile {
    network_plugin = "azure"
  }

  addon_profile {
    azure_policy {
      enabled = true
    }

    http_application_routing {
      enabled = false
    }

    kube_dashboard {
      enabled = false
    }

    oms_agent {
      enabled = false
      //log_analytics_workspace_id = azurerm_log_analytics_workspace.awp.id
    }
  }
}

resource "null_resource" "kubectl" {
  triggers = {
    always_run = timestamp()
  }

  provisioner "local-exec" {
    command     = "echo '${azurerm_kubernetes_cluster.cluster.kube_config_raw}' > ${var.kube_config_filepath}"
    interpreter = ["/bin/bash", "-c"]
  }
}

resource "null_resource" "aad-pod-identity" {
  depends_on = [null_resource.kubectl]

  provisioner "local-exec" {
    command     = "kubectl apply --kubeconfig=${var.kube_config_filepath} -f https://raw.githubusercontent.com/Azure/aad-pod-identity/master/deploy/infra/deployment-rbac.yaml"
    interpreter = ["/bin/bash", "-c"]
  }
}

output "client_certificate" {
  value = azurerm_kubernetes_cluster.cluster.kube_config.0.client_certificate
}

output "kube_config" {
  value = azurerm_kubernetes_cluster.cluster.kube_config_raw
}
