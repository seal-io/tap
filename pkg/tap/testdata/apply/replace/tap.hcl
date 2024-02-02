tap {
  path_syntax = "json_pointer"
}

resource "kubernetes_deployment" {
  type_alias = ["kubernetes_deployment_v1"]

  #
  # add
  #

  add {
    # to nested object attribute
    path  = "/metadata/0/labels/nested_object/key"
    value = "true"
  }

  add {
    # to nested object attribute
    path  = "/metadata/0/labels/nested_object/array_key"
    value = ["x", "y"]
  }

  add {
    # to nested array attribute
    path  = "/metadata/0/labels/nested_array/-1"
    value = "y"
  }

  add {
    # to nested array object attribute
    path  = "/metadata/0/labels/nested_array_object/1/0/key"
    value = "true"
  }

  add {
    # to nested object object attribute
    path  = "/metadata/0/labels/nested_object_object/x/y/key"
    value = "true"
  }

  add {
    # to attribute
    path  = "/metadata/0/labels/deploy"
    value = "true"
  }

  add {
    # to block
    path = "/spec/0/selector/0"
    value {
      match_labels = local.selectors
    }
  }

  add {
    # to block
    path = "/spec/0/template/0"
    value {
      metadata {
        labels = local.selectors
      }
    }
  }

  #
  # replace
  #

  replace {
    # in nested object attribute
    path  = "/metadata/0/labels/nested_object/key"
    value = "false"
  }

  replace {
    # in nested object attribute
    path  = "/metadata/0/labels/nested_object/array_key"
    value = ["x", "y", "z"]
  }

  replace {
    # in nested array attribute
    path  = "/metadata/0/labels/nested_array/-1"
    value = "z"
  }

  replace {
    # in nested array object attribute
    path  = "/metadata/0/labels/nested_array_object/1/0/key"
    value = "false"
  }

  replace {
    # in nested object object attribute
    path  = "/metadata/0/labels/nested_object_object/x/y/key"
    value = "false"
  }

  replace {
    # in attribute
    path  = "/metadata/0/labels/deploy"
    value = "false"
  }

  replace {
    # in block
    path = "/spec/0/selector/0"
    value {
      match_labels = merge(local.selectors, { "foo" = "bar" })
    }
  }

  replace {
    # in block
    path = "/spec/0/template/0"
    value {
      metadata {
        labels = merge(local.selectors, { "foo" = "bar" })
      }
    }
  }
}

resource "kubernetes_service" {
  type_alias = ["kubernetes_service_v1"]

  #
  # add
  #

  add {
    # to block
    path = "/spec/0"
    value {
      port {
        port        = 443
        target_port = 443
      }
    }
  }

  #
  # replace
  #

  replace {
    # in attribute
    path  = "/spec/0/port/0/port"
    value = 8080
  }

  replace {
    # in attribute
    path = "/spec/0/port/0"
    value {
      port        = 8080
      target_port = 8080
    }
  }
}
