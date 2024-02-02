tap {
  path_syntax = "json_pointer"
}

resource "kubernetes_deployment" {
  type_alias = ["kubernetes_deployment_v1"]

  #
  # set
  #

  set {
    # to nested object attribute
    path  = "/metadata/0/labels/nested_object/key"
    value = "true"
  }

  set {
    # to nested object attribute
    path  = "/metadata/0/labels/nested_object/array_key"
    value = ["x", "y"]
  }

  set {
    # to nested array attribute
    path  = "/metadata/0/labels/nested_array/-1"
    value = "y"
  }

  set {
    # to nested array object attribute
    path  = "/metadata/0/labels/nested_array_object/1/0/key"
    value = "true"
  }

  set {
    # to nested object object attribute
    path  = "/metadata/0/labels/nested_object_object/x/y/key"
    value = "true"
  }

  set {
    # to attribute
    path  = "/metadata/0/labels/deploy"
    value = "true"
  }

  set {
    # to block
    path = "/spec/0/selector/0"
    value {
      match_labels = local.selectors
    }
  }

  set {
    # to block
    path = "/spec/0/template/0"
    value {
      metadata {
        labels = local.selectors
      }
    }
  }
}

resource "kubernetes_service" {
  type_alias = ["kubernetes_service_v1"]

  #
  # set
  #

  set {
    # to block
    path = "/spec/0"
    value {
      port {
        port        = 443
        target_port = 443
      }
    }
  }
}
