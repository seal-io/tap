package terraform

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"

	"github.com/hashicorp/terraform/configs"
)

// mergeOverrides applies the overrides into the given configuration.
func mergeOverrides(cfg *Config) (*Config, error) {
	var err error

	rm := cfg.Root.Module

	if ress := rm.ManagedResources; ress != nil {
		ress, err = mergeResources(ress)
		if err != nil {
			return nil, fmt.Errorf("error applying managed resources: %w", err)
		}
		rm.ManagedResources = ress
	}

	if ress := rm.DataResources; ress != nil {
		ress, err = mergeResources(ress)
		if err != nil {
			return nil, fmt.Errorf("error applying data resources: %w", err)
		}
		rm.DataResources = ress
	}

	return cfg, nil
}

func mergeResources(ress map[string]*configs.Resource) (map[string]*configs.Resource, error) {
	for rn, r := range ress {
		if r.Config == nil {
			continue
		}

		var diags hcl.Diagnostics

		r.Config, diags = mergeBody(r.Config)
		if diags.HasErrors() {
			return nil, fmt.Errorf("error merging resource %q: %w", rn, diags)
		}

		ress[rn] = r
	}

	return ress, nil
}

func mergeBody(body hcl.Body) (hcl.Body, hcl.Diagnostics) {
	var b *configs.MergeBody

	switch t := body.(type) {
	default:
		return body, nil
	case *configs.MergeBody:
		b = t
	case configs.MergeBody:
		b = &t
	}

	bbo, err := mergeBody(b.Base)
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to merge body",
				Subject:  b.Base.MissingItemRange().Ptr(),
			},
		}
	}

	bb, ok := bbo.(*hclsyntax.Body)
	if !ok {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Invalid merge body",
				Detail:   fmt.Sprintf("Expected *hclsyntax.Body, got %T", b.Base),
				Subject:  b.Base.MissingItemRange().Ptr(),
			},
		}
	}

	obo, err := mergeBody(b.Override)
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to merge body",
				Subject:  b.Override.MissingItemRange().Ptr(),
			},
		}
	}

	ob, ok := obo.(*hclsyntax.Body)
	if !ok {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Invalid merge body",
				Detail:   fmt.Sprintf("Expected *hclsyntax.Body, got %T", b.Override),
				Subject:  b.Override.MissingItemRange().Ptr(),
			},
		}
	}

	// Merge attributes.
	for n, attr := range ob.Attributes {
		bb.Attributes[n] = attr
	}

	// Merge blocks.
	var (
		obts = make(map[string]bool)
		blks = make([]*hclsyntax.Block, 0, len(bb.Blocks))
	)

	for _, blk := range ob.Blocks {
		if blk.Type == "dynamic" {
			obts[blk.Labels[0]] = true
			continue
		}
		obts[blk.Type] = true
	}

	for _, blk := range bb.Blocks {
		if blk.Type == "dynamic" && obts[blk.Labels[0]] {
			continue
		}

		if obts[blk.Type] {
			continue
		}

		blks = append(blks, blk)
	}

	blks = append(blks, ob.Blocks...)
	bb.Blocks = blks

	return bb, nil
}
