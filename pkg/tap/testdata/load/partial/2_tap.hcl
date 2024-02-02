resource "kubernetes_deployment" {
  type_alias = ["kubernetes_deployment_v1"]
  name_match = ["nginx"]

  # always set.
  set {
    path  = ".metadata[0].namespace"
    value = "test"
  }

  # always set.
  set {
    path  = ".spec[0].template[0].spec[0].replicas"
    value = 2
  }

  # add if not exists.
  add {
    path  = ".spec[0].template[0].spec[0].containers[0].ports[?(@.container_port==8080 && @.protocol==TCP)]"
    value = {
      "container_port" = 80
      "protocol"       = "TCP"
    }
  }
}
