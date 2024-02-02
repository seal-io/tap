terraform {
  required_version = ">= 1.0"

  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.23.0"
    }
  }
}

variable "name" {
  type = string
}

variable "namespace" {
  type = string
}

locals {
  port      = 80
  selectors = {
    app = var.name
  }
}

resource "kubernetes_deployment_v1" "deploy" {
  metadata {
    name      = var.name
    namespace = var.namespace
  }
  spec {
    replicas = 1
    selector {
      match_labels = local.selectors
    }
    template {
      metadata {
        labels = local.selectors
      }
      spec {
        container {
          name  = var.name
          image = "nginx"
          port {
            container_port = local.port
          }
        }
      }
    }
  }
}

output "selectors" {
  value = local.selectors
}

output "port" {
  value = local.port
}
