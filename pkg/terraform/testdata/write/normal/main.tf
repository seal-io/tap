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
  selectors = {
    app = local.name
  }
}

resource "kubernetes_deployment_v1" "deploy" {
  metadata {
    name      = local.name
    namespace = local.namespace
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
          name  = "nginx"
          image = "nginx"
          port {
            container_port = 80
          }
        }
      }
    }
  }
}

resource "kubernetes_service_v1" "svc" {
  metadata {
    name      = kubernetes_deployment_v1.deploy.metadata[0].name
    namespace = kubernetes_deployment_v1.deploy.metadata[0].namespace
    labels = {
      usage           = "${local.name}-svc"
      binary_op_tmpl  = "%{if local.name == "nginx"}%%{}${local.namespace}%{endif}"
      ternary_op_tmpl = "%{if local.name == "nginx"}${local.namespace}%{else}default%{endif}"
    }
  }
  spec {
    selector = local.selectors
    type     = "ClusterIP"
    port {
      port        = 80
      target_port = 80
    }
  }
}

