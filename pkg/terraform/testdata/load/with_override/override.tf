# override other _override files.

terraform {
  cloud {
    organization = "example_corp"
    hostname     = "app.terraform.io"
    workspaces {}
  }
}

resource "kubernetes_service_v1" "svc" {
  # override dynamic block.
  spec {
    selector = local.selectors
    type     = "ClusterIP"
    port {
      port        = 80
      target_port = 80
    }
  }
}

resource "kubernetes_deployment_v1" "deploy" {
  # override static block.
  metadata {
    name      = "override"
    namespace = "override"
  }
}

resource "kubernetes_config_map_v1" "config" {
  # override meta argument.
  count = 1
}
