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

#  set {
#    path = ".metadata[0]"
#    value = {
#      name = "test"
#    }
#  }
#
#  set {
#    path = "."
#    value {
#      metadata {
#        name = "test"
#      }
#    }
#  }

  # remove if exists.
  remove {
    path = ".metadata[0].labels[1]"
  }
}

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
