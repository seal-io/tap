package terraform

import (
	"fmt"

	"github.com/hashicorp/go-version"

	"github.com/hashicorp/hcl/v2"

	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/modsdir"
)

// getWalker returns a module walker function that will load modules.
func getWalker(manifest modsdir.Manifest, parser *configs.Parser) configs.ModuleWalkerFunc {
	if len(manifest) == 0 {
		return nullWalker
	}

	// Copy from https://github.com/hashicorp/terraform/blob/v1.5.7/internal/configs/configload/loader_load.go.
	return func(req *configs.ModuleRequest) (*configs.Module, *version.Version, hcl.Diagnostics) {
		// Since we're just loading here, we expect that all referenced modules
		// will be already installed and described in our manifest. However, we
		// do verify that the manifest and the configuration are in agreement
		// so that we can prompt the user to run "terraform init" if not.
		key := manifest.ModuleKey(req.Path)

		record, exists := manifest[key]
		if !exists {
			return nil, nil, hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary:  "Module not installed",
					Detail:   "This module is not yet installed. Run \"terraform init\" to install all modules required by this configuration.",
					Subject:  &req.CallRange,
				},
			}
		}

		var wDiags hcl.Diagnostics

		// Check for inconsistencies between manifest and config.

		// We ignore a nil SourceAddr here, which represents a failure during
		// configuration parsing, and will be reported in a diagnostic elsewhere.
		if req.SourceAddr != nil && req.SourceAddr.String() != record.SourceAddr {
			wDiags = append(wDiags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Module source has changed",
				Detail: "The source address was changed since this module was installed. " +
					"Run \"terraform init\" to install all modules required by this configuration.",
				Subject: &req.SourceAddrRange,
			})
		}

		if len(req.VersionConstraint.Required) > 0 && record.Version == nil {
			wDiags = append(wDiags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Module version requirements have changed",
				Detail: "The version requirements have changed since this module was installed " +
					"and the installed version is no longer acceptable." +
					" Run \"terraform init\" to install all modules required by this configuration.",
				Subject: &req.SourceAddrRange,
			})
		}

		if record.Version != nil && !req.VersionConstraint.Required.Check(record.Version) {
			wDiags = append(wDiags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Module version requirements have changed",
				Detail: fmt.Sprintf(
					"The version requirements have changed since this module was installed "+
						"and the installed version (%s) is no longer acceptable. "+
						"Run \"terraform init\" to install all modules required by this configuration.",
					record.Version,
				),
				Subject: &req.SourceAddrRange,
			})
		}

		mod, mDiags := parser.LoadConfigDir(record.Dir)
		if mDiags.HasErrors() {
			wDiags = wDiags.Extend(mDiags)
		}

		if mod == nil {
			// Nil specifically indicates that the directory does not exist or
			// cannot be read, so in this case we'll discard any generic diagnostics
			// returned from LoadConfigDir and produce our own context-sensitive
			// error message.
			return nil, nil, hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary:  "Module not installed",
					Detail: fmt.Sprintf(
						"This module's local cache directory %s could not be read. Run \"terraform init\" to install all modules required by this configuration.",
						record.Dir,
					),
					Subject: &req.CallRange,
				},
			}
		}

		return mod, record.Version, wDiags
	}
}

func nullWalker(_ *configs.ModuleRequest) (*configs.Module, *version.Version, hcl.Diagnostics) {
	return nil, nil, nil
}
