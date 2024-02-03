# Terraform Advanced Patcher (TAP)

[![](https://goreportcard.com/badge/github.com/seal-io/tap)](https://goreportcard.com/report/github.com/seal-io/tap)
[![](https://img.shields.io/github/actions/workflow/status/seal-io/tap/ci.yml?label=ci)](https://github.com/seal-io/tap/actions)
[![](https://img.shields.io/github/v/tag/seal-io/tap?label=release)](https://github.com/seal-io/tap/releases)
[![](https://img.shields.io/github/downloads/seal-io/tap/total)](https://github.com/seal-io/tap/releases)
[![](https://img.shields.io/github/license/seal-io/tap?label=license)](https://github.com/seal-io/tap#license)

Terraform Advanced Patcher, aka. **TAP**, is a tool to patch [Terraform](https://www.terraform.io/) file.

This tool is maintained by [Seal](https://github.com/seal-io).

## Background

In some cases, consuming native [Terraform Override](https://developer.hashicorp.com/terraform/language/files/override)
can complete some additional expansion, like overriding a nested block or changing a predefined attribute, but the
capabilities are limited. For example, it needs accurate block header to patch, and it's impossible to conditionally
make changes to nested blocks or attributes.

**TAP** is designed to satisfy the above features.

## Implementation

As we all know, the HCL used
by [Terraform supports JSON syntax](https://github.com/hashicorp/hcl/blob/main/json/spec.md). Therefore, **TAP** can be
implemented using JSON patching.

The patching mode of [Terraform Override](https://developer.hashicorp.com/terraform/language/files/override) looks
like [JSON Merge Patch, RFC 7386](https://datatracker.ietf.org/doc/html/rfc7386), but **TAP** is working
as [JSON Patch, RFC 6902](https://datatracker.ietf.org/doc/html/rfc6902).
> For a comparison of JSON patch and JSON merge patch,
> see [JSON Patch and JSON Merge Patch](https://erosb.github.io/post/json-patch-vs-merge-patch/).

### Note

Although **TAP** and **Terraform Override** path in different ways, they have one thing in same, that is, top-level
blocks cannot be deleted. The core reason is that top-level blocks can
configure [Meta-Argument](https://developer.hashicorp.com/terraform/language/meta-arguments/count) or participate in the
configuration of [Meta-Argument](https://developer.hashicorp.com/terraform/language/meta-arguments/depends_on). Directly
deleting a top-level block will cause many problems.

## Usage

**TAP** is not a complete JSON patch
for [Terraform JSON Configuration Syntax](https://developer.hashicorp.com/terraform/language/syntax/json)The original
JSON path needs its operation to have exactly one "path" member, which values with
a [JSON Pointer](https://datatracker.ietf.org/doc/html/rfc6901), but **TAP** implements limited
operations: `add`, `remove`, and `replace`, and also introduces a new operation: `set`.

```hcl
# tap.hcl

tap {
  path_syntax = "json_pointer"
}

resource "kubernetes_deployment" {
  type_alias = ["kubernetes_deployment_v1"]

  add {
    path  = "/metadata/0/labels/new-label"
    value = "new-label-value"
  }

  remove {
    path = "/metadata/0/labels/old-label"
  }

  replace {
    path  = "/spec/0/template/0/spec/0/replicas"
    value = "2"
  }

  set {
    path = "/spec/0/selector/0"
    value {
      match_labels = local.selectors
    }
  }
}
```

> **TAP** recognizes the path syntax according to the `path_syntax` attribute in the `tap` block, in which the default
> value is `json_pointer`. We are going to support more path syntax in the future.

**TAP**, at present, only supports patching `resource` and `data` blocks, and filters out the target blocks
by `type_alias` or `name_match` attributes.

```hcl
# tap.hcl

tap {
  path_syntax = "json_pointer"
}

resource "kubernetes_deployment" {
  type_alias = ["kubernetes_deployment_v1"]
  name_match = ["nginx"]

  # ... operations
}

data "kubernetes_config_map" {
  type_alias = ["kubernetes_config_map_v1"]
  name_match = ["nginx"]

  # ... operations
}
```

**TAP** also allows ignoring error if patching fails, and it can be configured by the `continue_on_error` attribute in
the `tap` block.

```hcl
# tap.hcl

tap {
  continue_on_error = true               # global
  path_syntax       = "json_pointer"
}

resource "kubernetes_deployment" {
  continue_on_error = false              # local
  type_alias        = ["kubernetes_deployment_v1"]

  # ... operations
}
```

**TAP** is a wrapper to [Terraform](https://www.terraform.io/) or [OpenTofu](https://opentofu.org/), you can simplify
alias **TAP** as `tf`, and use it as a drop-in replacement for Terraform or OpenTofu.
> **TAP** is not a fork of Terraform or OpenTofu, so you still need to install the CLI
> of [Terraform](https://developer.hashicorp.com/terraform/install)
> or [OpenTofu](https://opentofu.org/docs/intro/install)
> at first.

```bash
$ go install github.com/seal-io/tap/cmd/tap@latest
$ mv "${GOPATH}"/bin/tap "${GOPATH}"/bin/tf
$ tf --version
```

Put the `tap.hcl` file in the same directory as the `main.tf` file, and then execute `tf plan` or `tf apply` to see the
effect.

### Example YouTube Overview

[![](https://img.youtube.com/vi/hk-uvKwsDPs/maxresdefault.jpg)](https://www.youtube.com/watch?v=hk-uvKwsDPs)

## Requirements

- [Go](https://golang.org/doc/install) 1.20+ (to build the provider plugin)
- [Terraform](https://www.terraform.io/) v1.5+ or [OpenTofu](https://opentofu.org/) v1.6+ (to run with **TAP**)

# License

[Mozilla Public License v2.0](./LICENSE)
