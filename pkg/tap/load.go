package tap

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/afero"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"

	"github.com/seal-io/tap/utils/pointer"
)

type (
	Config struct {
		PathSyntax string
		Patches    []Patch
	}

	Patch struct {
		ContinueOnError bool
		ResourceMode    string // Select from "resource" or "data".
		ResourceTypes   []string
		ResourceNames   []string
		Operations      []Operation
	}

	Operation struct {
		Mode  string // Select from "add", "remove", "replace", or "set".
		Path  string
		Value Value
	}

	Value struct {
		Attribute *hcl.Attribute
		Block     *hcl.Block
	}
)

// HasConfig checks if the given directory has a tap configuration.
func HasConfig(dir string) (bool, error) {
	fs := afero.Afero{Fs: afero.NewBasePathFs(afero.NewOsFs(), dir)}

	{
		gs, err := afero.Glob(fs, "*_tap.hcl")
		if err != nil {
			return false, fmt.Errorf("failed to glob files suffix with `_tap.hcl`: %w", err)
		}

		if len(gs) > 0 {
			return true, nil
		}
	}

	{
		si, err := fs.Stat("tap.hcl")
		if err != nil && !os.IsNotExist(err) {
			return false, fmt.Errorf("failed to stat file `tap.hcl`: %w", err)
		}

		if si != nil && !si.IsDir() {
			return true, nil
		}
	}

	return false, nil
}

// Load loads the tap configuration from the given directory,
// returns nil if no tap configuration is found.
func Load(dir string) (*Config, error) {
	fs := afero.Afero{Fs: afero.NewBasePathFs(afero.NewOsFs(), dir)}

	var files []os.FileInfo
	{
		gs, err := afero.Glob(fs, "*_tap.hcl")
		if err != nil {
			return nil, fmt.Errorf("failed to glob files suffix with `_tap.hcl`: %w", err)
		}

		files = make([]os.FileInfo, 0, len(gs))

		for i := range gs {
			si, err := fs.Stat(gs[i])
			if err != nil && !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to stat file `%s`: %w", gs[i], err)
			}

			if si != nil && !si.IsDir() {
				files = append(files, si)
			}
		}

		si, err := fs.Stat("tap.hcl")
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to stat file `tap.hcl`: %w", err)
		}

		if si != nil && !si.IsDir() {
			files = append(files, si)
		}
	}

	if len(files) == 0 {
		return nil, nil
	}

	bodies := make([]*hcl.File, 0, len(files))
	{
		parser := hclparse.NewParser()

		for i := range files {
			fn := files[i].Name()

			bs, err := fs.ReadFile(fn)
			if err != nil {
				return nil, fmt.Errorf("failed to read file `%s`: %w", fn, err)
			}

			f, d := parser.ParseHCL(bs, fn)
			if d.HasErrors() {
				return nil, fmt.Errorf("failed to parse file `%s`: %w", fn, d)
			}

			bodies = append(bodies, f)
		}
	}

	cfg, diags := buildConfig(hcl.MergeFiles(bodies))
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to build tap config: %w", diags)
	}

	return cfg, nil
}

func buildConfig(body hcl.Body) (*Config, hcl.Diagnostics) {
	bc, remain, diags := body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type: "tap",
			},
		},
	})
	if diags.HasErrors() {
		return nil, diags
	}

	b := bc.Blocks.OfType("tap")
	if b == nil {
		return nil, diags
	}

	var v struct {
		ContinueOnError bool     `hcl:"continue_on_error,optional"`
		PathSyntax      string   `hcl:"path_syntax,optional"`
		Remain          hcl.Body `hcl:",remain"`
	}

	diags = gohcl.DecodeBody(b[0].Body, nil, &v)
	if diags.HasErrors() {
		return nil, diags
	}

	if v.PathSyntax == "" {
		v.PathSyntax = "json_pointer"
	}

	cfg := Config{
		PathSyntax: v.PathSyntax,
	}
	diags = buildResourcePatches(remain, &cfg, v.ContinueOnError)

	return &cfg, diags
}

func buildResourcePatches(remain hcl.Body, cfg *Config, coe bool) hcl.Diagnostics {
	bc, diags := remain.Content(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "resource",
				LabelNames: []string{"type"},
			},
			{
				Type:       "data",
				LabelNames: []string{"type"},
			},
		},
	})
	if diags.HasErrors() {
		return diags
	}

	if cfg.Patches == nil {
		cfg.Patches = make([]Patch, 0, len(bc.Blocks))
	}

	for i := range bc.Blocks {
		b := bc.Blocks[i]

		var v struct {
			ContinueOnError *bool    `hcl:"continue_on_error,optional"`
			TypeAlias       []string `hcl:"type_alias,optional"`
			NameMatch       []string `hcl:"name_match,optional"`
			Remain          hcl.Body `hcl:",remain"`
		}

		dDiags := gohcl.DecodeBody(b.Body, nil, &v)
		if dDiags.HasErrors() {
			diags = diags.Extend(dDiags)
			continue
		}

		rp := Patch{
			ContinueOnError: pointer.BoolDeref(v.ContinueOnError, coe),
			ResourceMode:    b.Type,
			ResourceTypes:   append([]string{b.Labels[0]}, v.TypeAlias...),
			ResourceNames:   v.NameMatch,
		}

		dDiags = buildOperations(v.Remain, &rp)
		if dDiags.HasErrors() {
			diags = diags.Extend(dDiags)
			continue
		}

		if len(rp.Operations) == 0 {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary: fmt.Sprintf(
					"Patch %q block requires at least one Operation block",
					strings.Join([]string{rp.ResourceMode, rp.ResourceTypes[0]}, " ")),
				Subject: pointer.Ref(v.Remain.MissingItemRange()),
			})
		}

		cfg.Patches = append(cfg.Patches, rp)
	}

	return diags
}

func buildOperations(remain hcl.Body, rp *Patch) hcl.Diagnostics {
	bc, diags := remain.Content(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			// Basic mode.
			{
				Type: "add",
			},
			{
				Type: "remove",
			},
			{
				Type: "replace",
			},
			// Aggregate mode.
			{
				Type: "set",
			},
		},
	})
	if diags.HasErrors() {
		return diags
	}

	if rp.Operations == nil {
		rp.Operations = make([]Operation, 0, len(bc.Blocks))
	}

	for i := range bc.Blocks {
		b := bc.Blocks[i]

		var v struct {
			Path   string   `hcl:"path"`
			Remain hcl.Body `hcl:",remain"`
		}

		dDiags := gohcl.DecodeBody(b.Body, nil, &v)
		if dDiags.HasErrors() {
			diags = diags.Extend(dDiags)
			continue
		}

		op := Operation{
			Mode: b.Type,
			Path: v.Path,
		}

		if b.Type != "remove" {
			dDiags = buildValue(v.Remain, &op)
			if dDiags.HasErrors() {
				diags = diags.Extend(dDiags)
				continue
			}
		}

		rp.Operations = append(rp.Operations, op)
	}

	return diags
}

func buildValue(remain hcl.Body, op *Operation) hcl.Diagnostics {
	bc, diags := remain.Content(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name: "value",
			},
		},
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type: "value",
			},
		},
	})
	if diags.HasErrors() {
		return diags
	}

	if attr, exist := bc.Attributes["value"]; exist {
		op.Value.Attribute = attr
		return nil
	} else if blks := bc.Blocks.OfType("value"); len(blks) == 1 {
		op.Value.Block = blks[0]
		return nil
	}

	return hcl.Diagnostics{
		{
			Severity: hcl.DiagError,
			Summary: fmt.Sprintf(
				"Operation %q block requires either a value attribute or a value block",
				op.Mode),
			Subject: &bc.MissingItemRange,
		},
	}
}
