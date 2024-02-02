package terraform

import (
	"io"
	"sort"

	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/terraform/configs"
)

// Write writes the given Config to the given writer.
func Write(cfg *Config, writer io.Writer) error {
	var (
		m  = cfg.Root.Module
		wf = hclwrite.NewFile()
	)

	// block: terraform
	{
		wb := wf.Body()
		tfBody := wb.AppendNewBlock("terraform", nil).Body()

		// attribute: required_version
		if len(m.CoreVersionConstraints) != 0 && len(m.CoreVersionConstraints[0].Required) != 0 {
			tfBody.SetAttributeValue("required_version",
				cty.StringVal(m.CoreVersionConstraints[0].Required.String()))
			tfBody.AppendNewline()
		}

		// attribute: experiments
		if len(m.ActiveExperiments) != 0 {
			eBody := make([]cty.Value, 0, len(m.ActiveExperiments))
			for k := range m.ActiveExperiments {
				eBody = append(eBody, cty.StringVal(k.Keyword()))
			}

			tfBody.SetAttributeValue("experiments", cty.ListVal(eBody))
			tfBody.AppendNewline()
		}

		switch {
		case m.Backend != nil:
			// block: backend
			bgBody := tfBody.AppendNewBlock("backend", []string{m.Backend.Type}).Body()
			bgBody.AppendHCLBody(m.Backend.Config)
			tfBody.AppendNewline()
		case m.CloudConfig != nil:
			// block: cloud
			ccBody := tfBody.AppendNewBlock("cloud", nil).Body()
			ccBody.AppendHCLBody(m.CloudConfig.Config)
			tfBody.AppendNewline()
		}

		// block: required_providers
		if m.ProviderRequirements != nil && len(m.ProviderRequirements.RequiredProviders) != 0 {
			requiredProviders := make([]*configs.RequiredProvider, 0, len(m.ProviderRequirements.RequiredProviders))
			for k := range m.ProviderRequirements.RequiredProviders {
				requiredProviders = append(requiredProviders, m.ProviderRequirements.RequiredProviders[k])
			}

			sort.Slice(requiredProviders, func(i, j int) bool {
				if requiredProviders[i].DeclRange.Filename == requiredProviders[j].DeclRange.Filename {
					return requiredProviders[i].DeclRange.Start.Line < requiredProviders[j].DeclRange.Start.Line
				}

				return requiredProviders[i].DeclRange.Filename < requiredProviders[j].DeclRange.Filename
			})

			rpBody := tfBody.AppendNewBlock("required_providers", nil).Body()

			for i, rp := range requiredProviders {
				rpCtyValMap := map[string]cty.Value{}
				if rp.Source != "" {
					rpCtyValMap["source"] = cty.StringVal(rp.Source)
				}

				if rp.Requirement.Required.String() != "" {
					rpCtyValMap["version"] = cty.StringVal(rp.Requirement.Required.String())
				}

				if len(rpCtyValMap) == 0 {
					continue
				}

				rpBody.SetAttributeValue(rp.Name, cty.ObjectVal(rpCtyValMap))

				if i != len(requiredProviders)-1 {
					rpBody.AppendNewline()
				}
			}
		}

		// block: provider_meta
		if len(m.ProviderMetas) != 0 {
			for _, pm := range m.ProviderMetas {
				pmBody := tfBody.AppendNewBlock("provider_meta", []string{pm.Provider}).Body()
				pmBody.AppendHCLBody(pm.Config)
			}

			tfBody.AppendNewline()
		}

		wb.AppendNewline()
	}

	// block: provider
	{
		providers := make([]*configs.Provider, 0, len(m.ProviderConfigs))
		for k := range m.ProviderConfigs {
			providers = append(providers, m.ProviderConfigs[k])
		}

		sort.Slice(providers, func(i, j int) bool {
			if providers[i].DeclRange.Filename == providers[j].DeclRange.Filename {
				return providers[i].DeclRange.Start.Line < providers[j].DeclRange.Start.Line
			}

			return providers[i].DeclRange.Filename < providers[j].DeclRange.Filename
		})

		wb := wf.Body()
		for _, p := range providers {
			pBody := wb.AppendNewBlock("provider", []string{p.Name}).Body()
			if p.Alias != "" {
				pBody.SetAttributeValue("alias", cty.StringVal(p.Alias))
			}

			if len(p.Version.Required) != 0 {
				pBody.SetAttributeValue("version", cty.StringVal(p.Version.Required.String()))
			}

			pBody.AppendHCLBody(p.Config)
			wb.AppendNewline()
		}
	}

	// block: variable
	{
		variables := make([]*configs.Variable, 0, len(m.Variables))
		for k := range m.Variables {
			variables = append(variables, m.Variables[k])
		}

		sort.Slice(variables, func(i, j int) bool {
			if variables[i].DeclRange.Filename == variables[j].DeclRange.Filename {
				return variables[i].DeclRange.Start.Line < variables[j].DeclRange.Start.Line
			}

			return variables[i].DeclRange.Filename < variables[j].DeclRange.Filename
		})

		wb := wf.Body()
		for _, v := range variables {
			wb.AppendUnstructuredTokens(hclwrite.Tokens{
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte("variable"),
				},
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte(`"` + v.Name + `"`),
				},
			})
			wb.AppendUnstructuredTokens(fromHCLTokens(v.Tokens, false))
			wb.AppendNewline()
		}
	}

	// block: local
	{
		locals := make([]*configs.Local, 0, len(m.Locals))
		for k := range m.Locals {
			locals = append(locals, m.Locals[k])
		}

		sort.Slice(locals, func(i, j int) bool {
			if locals[i].DeclRange.Filename == locals[j].DeclRange.Filename {
				return locals[i].DeclRange.Start.Line < locals[j].DeclRange.Start.Line
			}

			return locals[i].DeclRange.Filename < locals[j].DeclRange.Filename
		})

		if len(locals) != 0 {
			wb := wf.Body()

			lsBody := wb.AppendNewBlock("locals", nil).Body()
			for _, l := range locals {
				lsBody.SetAttributeRaw(l.Name, fromHCLTokens(l.Tokens, true))
			}

			wb.AppendNewline()
		}
	}

	// block: resource
	{
		resources := make([]*configs.Resource, 0, len(m.ManagedResources))
		for k := range m.ManagedResources {
			resources = append(resources, m.ManagedResources[k])
		}

		sort.Slice(resources, func(i, j int) bool {
			if resources[i].DeclRange.Filename == resources[j].DeclRange.Filename {
				return resources[i].DeclRange.Start.Line < resources[j].DeclRange.Start.Line
			}

			return resources[i].DeclRange.Filename < resources[j].DeclRange.Filename
		})

		wb := wf.Body()
		for _, r := range resources {
			rBody := wb.AppendNewBlock("resource", []string{r.Type, r.Name}).Body()

			// attribute: provider
			if r.ProviderConfigRef != nil {
				rBody.SetAttributeValue("provider", cty.StringVal(r.ProviderConfigRef.Name))
				rBody.AppendNewline()
			}

			rBody.AppendHCLBody(r.Config)
			wb.AppendNewline()
		}
	}

	// block: data
	{
		datas := make([]*configs.Resource, 0, len(m.DataResources))
		for k := range m.DataResources {
			datas = append(datas, m.DataResources[k])
		}

		sort.Slice(datas, func(i, j int) bool {
			if datas[i].DeclRange.Filename == datas[j].DeclRange.Filename {
				return datas[i].DeclRange.Start.Line < datas[j].DeclRange.Start.Line
			}

			return datas[i].DeclRange.Filename < datas[j].DeclRange.Filename
		})

		wb := wf.Body()
		for _, d := range datas {
			dBody := wb.AppendNewBlock("data", []string{d.Type, d.Name}).Body()

			// attribute: provider
			if d.ProviderConfigRef != nil {
				dBody.SetAttributeValue("provider", cty.StringVal(d.ProviderConfigRef.Name))
				dBody.AppendNewline()
			}

			dBody.AppendHCLBody(d.Config)
			wb.AppendNewline()
		}
	}

	// block: module
	{
		modules := make([]*configs.ModuleCall, 0, len(m.ModuleCalls))
		for k := range m.ModuleCalls {
			modules = append(modules, m.ModuleCalls[k])
		}

		sort.Slice(modules, func(i, j int) bool {
			if modules[i].DeclRange.Filename == modules[j].DeclRange.Filename {
				return modules[i].DeclRange.Start.Line < modules[j].DeclRange.Start.Line
			}

			return modules[i].DeclRange.Filename < modules[j].DeclRange.Filename
		})

		wb := wf.Body()
		for _, m := range modules {
			mBody := wb.AppendNewBlock("module", []string{m.Name}).Body()

			// attribute: source
			if m.SourceSet {
				mBody.SetAttributeValue("source", cty.StringVal(m.SourceAddrRaw))
				mBody.AppendNewline()
			}

			// attribute: version
			if len(m.Version.Required) != 0 {
				mBody.SetAttributeValue("required_version", cty.StringVal(m.Version.Required.String()))
				mBody.AppendNewline()
			}

			mBody.AppendHCLBody(m.Config)
			wb.AppendNewline()
		}
	}

	// block: output
	{
		outputs := make([]*configs.Output, 0, len(m.Outputs))
		for k := range m.Outputs {
			outputs = append(outputs, m.Outputs[k])
		}

		sort.Slice(outputs, func(i, j int) bool {
			if outputs[i].DeclRange.Filename == outputs[j].DeclRange.Filename {
				return outputs[i].DeclRange.Start.Line < outputs[j].DeclRange.Start.Line
			}

			return outputs[i].DeclRange.Filename < outputs[j].DeclRange.Filename
		})

		wb := wf.Body()
		for i, o := range outputs {
			wb.AppendUnstructuredTokens(hclwrite.Tokens{
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte("output"),
				},
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte(`"` + o.Name + `"`),
				},
			})
			wb.AppendUnstructuredTokens(fromHCLTokens(o.Tokens, false))

			if i != len(outputs)-1 {
				wb.AppendNewline()
			}
		}
	}

	_, err := wf.WriteTo(writer)

	return err
}
