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
  # remove
  #

  remove {
    # from attribute
    path = "/metadata/0/labels/nested_object/key"
  }

  remove {
    # from attribute
    path = "/metadata/0/labels/nested_object/array_key"
  }

  remove {
    # from attribute
    path = "/metadata/0/labels/nested_array/-1"
  }

  remove {
    # from attribute
    path = "/metadata/0/labels/nested_array_object/1/0/key"
  }

  remove {
    # from attribute
    path = "/metadata/0/labels/nested_object_object/x/y/key"
  }

  remove {
    # from attribute
    path = "/metadata/0/labels/deploy"
  }

  remove {
    # from block
    path = "/spec/0/selector/0/match_labels"
  }

  remove {
    # from block
    path = "/spec/0/template/0/metadata"
  }
}
