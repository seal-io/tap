// Copyright (c) 2024 Seal, Inc.
// Copyright (c) 2014 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Package terraform borrows from the https://github.com/hashicorp/terraform/tree/v1.5.7.
//
// Since the default Terraform hasn't exposed its internal packages,
// this package shifts up the internal packages to be public.
//
// Secondary, this package exposes the configs.MergeBody type,
// which is used to merge the HCL bodies in the Terraform module,
// and exposes the AST tokens of some Terraform blocks.
//
// This package is not exactly the same with the original implementation but works well for us.
package terraform
