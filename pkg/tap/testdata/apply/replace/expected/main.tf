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
        key       = "false"
        array_key = ["x", "y", "z"]
      }
      nested_array = ["x", "z"]
      nested_array_object = [{
        x = "y"
        }, [{
          y   = "x"
          key = "false"
      }]]
      nested_object_object = {
        x = {
          y = {
            z   = "x"
            key = "false"
          }
        }
      }
      deploy = "false"
    }
  }
  spec {
    replicas = 1
    selector {
      match_labels = merge(local.selectors, {
        "foo" = "bar"
      })
    }
    template {
      metadata {
        labels = merge(local.selectors, {
          "foo" = "bar"
        })
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
      selector = local.selectors
      type     = "ClusterIP"
      port {
        port        = 8080
        target_port = 8080
      }
      port {
        port        = 443
        target_port = 443
      }
    }
  }
}

