terraform {
  required_version = ">= 1.0"

  required_providers {
    kubernetes = {
      source  = "fake/kubernetes"
      version = ">= 0.1.0"
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
    labels = {
      app = local.name
      nested_object = {
        app       = local.name
        key       = "true"
        array_key = ["x", "y"]
      }
      nested_array = ["y"]
      nested_array_object = [{
        x = "y"
        }, [{
          y   = "x"
          key = "true"
      }]]
      nested_object_object = {
        x = {
          y = {
            z   = "x"
            key = "true"
          }
        }
      }
      deploy = "true"
    }
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
    }
  }
}

resource "kubernetes_service_v1" "svc" {
  metadata {
    name      = kubernetes_deployment_v1.deploy.metadata[0].name
    namespace = kubernetes_deployment_v1.deploy.metadata[0].namespace
  }
  dynamic "spec" {
    for_each = [{}]
    content {
      port {
        port        = 443
        target_port = 443
      }
    }
  }
}

