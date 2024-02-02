tap {
  path_syntax = "json_pointer"
}

resource "kubernetes_deployment" {
  type_alias = ["kubernetes_deployment_v1"]

  set {
    path  = "/spec/0/replicas"
    value = 3
  }
}
