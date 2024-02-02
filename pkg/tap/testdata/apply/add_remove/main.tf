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
    labels    = {
      app           = local.name
      # # patch # deploy = "true"
      nested_object = {
        app = local.name
        # # patch # key = "true"
        # # patch # array_key = ["x", "y"]
      }
      nested_array = [
        "x"
        # # patch # "y"
      ]
      nested_array_object = [
        {
          x = "y"
        },
        [
          {
            y = "x"
            # # patch # key = "true"
          }
        ]
      ]
      nested_object_object = {
        x = {
          y = {
            z = "x"
            # # patch # key = "true"
          }
        }
      }
    }
  }
  spec {
    replicas = 1
    selector {
      # # patch # match_labels = local.selectors
    }
    template {
      # # patch #
      # metadata {
      #   labels = local.selectors
      # }
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
      # # patch #
      # port {
      #    port        = 443
      #    target_port = 443
      # }
    }
  }
}
