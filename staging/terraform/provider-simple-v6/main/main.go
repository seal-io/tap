// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"github.com/hashicorp/terraform/grpcwrap"
	plugin "github.com/hashicorp/terraform/plugin6"
	simple "github.com/hashicorp/terraform/provider-simple-v6"
	"github.com/hashicorp/terraform/tfplugin6"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		GRPCProviderFunc: func() tfplugin6.ProviderServer {
			return grpcwrap.Provider6(simple.Provider())
		},
	})
}
