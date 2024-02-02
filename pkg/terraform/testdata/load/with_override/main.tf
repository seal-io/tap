terraform {
  required_version = ">= 1.0"

  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.23.0"
    }
  }

  backend "http" {
    address = "http://myrest.api.com/foo"
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
  }

  dynamic "spec" {
    for_each = [{}]
    content {
      selector = local.selectors
      type     = "ClusterIP"
      port {
        port        = 80
        target_port = 80
      }
    }
  }
}

resource "kubernetes_config_map_v1" "config" {
  count = 2

  metadata {
    generate_name = format("%s-", kubernetes_deployment_v1.deploy.metadata[0].name)
    namespace     = kubernetes_deployment_v1.deploy.metadata[0].namespace
  }
  data = {
    "k1" = "v1"
    "k2" = "v2"
  }
}

resource "kubernetes_secret_v1" "secret" {
  for_each = [{}]

  metadata {
    generate_name = format("%s-", kubernetes_deployment_v1.deploy.metadata[0].name)
    namespace     = kubernetes_deployment_v1.deploy.metadata[0].namespace
  }
  data = {
    "k1" = "v1"
    "k2" = "v2"
  }
}
