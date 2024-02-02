resource "kubernetes_service" {
  type_alias = ["kubernetes_service_v1"]
  name_match = ["nginx", "nginx-svc"]

  # always set.
  set {
    path  = ".metadata[0].namespace"
    value = "test"
  }

  # add to the end of the array if ports array is found.
  add {
    path  = ".spec[0].ports[-1]"
    value = {
      "name"       = "http"
      "port"       = 80
      "targetPort" = 80
    }
  }

  # replace if exists.
  replace {
    path  = ".spec[0].type"
    value = "NodePort"
  }
}
