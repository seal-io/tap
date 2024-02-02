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
  count = local.namespace == "default" ? 1 : 0
  source = "./modules/nginx"

  name      = local.name
  namespace = local.namespace
}

resource "kubernetes_service_v1" "svc" {
  metadata {
    name      = local.name
    namespace = local.namespace
  }
  spec {
    selector = module.nginx[0].selectors
    type     = "ClusterIP"
    port {
      port        = module.nginx[0].port
      target_port = module.nginx[0].port
    }
  }
}
