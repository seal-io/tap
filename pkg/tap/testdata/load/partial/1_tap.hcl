tap {
  continue_on_error = true
  path_syntax       = "tap_pointer"
}

resource "kubernetes_namespace" {
  type_alias = ["kubernetes_namespace_v1"]
  name_match = null # match all namespaces.

  # always set.
  set {
    path  = ".metadata[0].name"
    value = "test"
  }

  # remove if exists.
  remove {
    path = ".metadata[0].labels[1]"
  }
}

data "kubernetes_namespace" {
  continue_on_error = false

  type_alias = ["kubernetes_namespace_v1"]
  name_match = ["nginx-svc"]

  # replace if exists.
  replace {
    path  = ".metadata[0].name"
    value = "test"
  }
}
