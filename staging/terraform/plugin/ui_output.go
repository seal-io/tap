// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package plugin

import (
	"net/rpc"

	"github.com/hashicorp/terraform/terraform"
)

// UIOutput is an implementatin of terraform.UIOutput that communicates
// over RPC.
type UIOutput struct {
	Client *rpc.Client
}

func (o *UIOutput) Output(v string) {
	o.Client.Call("Plugin.Output", v, new(any))
}

// UIOutputServer is the RPC server for serving UIOutput.
type UIOutputServer struct {
	UIOutput terraform.UIOutput
}

func (s *UIOutputServer) Output(
	v string,
	reply *any) error {
	s.UIOutput.Output(v)
	return nil
}
