terraform {
  required_version = ">= 1.0"

  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.23.0"
    }
  }
}

provider "kubernetes" {
  config_path = "~/.kube/config"
}

locals {
  namespace = "default"
  name      = "nginx"
}

module "nginx" {
  source = "modules/nginx"

  name      = local.name
  namespace = local.namespace
}

resource "kubernetes_service_v1" "svc" {
  metadata {
    name      = local.name
    namespace = local.namespace
  }
  spec {
    selector = module.nginx.selectors
    type     = "ClusterIP"
    port {
      port        = module.nginx.port
      target_port = module.nginx.port
    }
  }
}
