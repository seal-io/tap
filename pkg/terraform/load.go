package terraform

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/modsdir"
	"github.com/spf13/afero"
)

type Config = configs.Config

// Load loads the terraform configuration from the given directory,
// applying overrides to the configuration as necessary.
func Load(dir string) (*Config, error) {
	fs := afero.Afero{Fs: afero.NewBasePathFs(afero.NewOsFs(), dir)}

	parser := configs.NewParser(fs)
	parser.AllowLanguageExperiments(true)

	var walker configs.ModuleWalkerFunc
	{
		manifest := modsdir.Manifest{}
		f, err := fs.Open(filepath.Join(".terraform", "modules", modsdir.ManifestSnapshotFilename))

		if !errors.Is(err, os.ErrNotExist) {
			manifest, err = modsdir.ReadManifestSnapshot(f)
			_ = f.Close()

			if err != nil {
				return nil, fmt.Errorf("failed to read modules manifest: %w", err)
			}
		}

		walker = getWalker(manifest, parser)
	}

	root, diags := parser.LoadConfigDir(".")
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to load terraform root module: %w", diags)
	}

	if root == nil {
		return nil, nil
	}

	cfg, diags := configs.BuildConfig(root, walker)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to build terraform config: %w", diags)
	}

	cfg, err := mergeOverrides(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to apply overrides: %w", err)
	}

	return cfg, nil
}
