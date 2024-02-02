package tap

import (
	"fmt"

	"golang.org/x/exp/slices"

	"github.com/hashicorp/terraform/configs"
)

// Apply applies the tap configuration to the Terraform configuration.
func Apply(tfCfg *configs.Config, cfg *Config) (*configs.Config, error) {
	if cfg == nil {
		return tfCfg, nil
	}

	for i := range cfg.Patches {
		p := cfg.Patches[i]

		// Select typed resources.
		originalRess := tfCfg.Module.ManagedResources
		if p.ResourceMode == "data" {
			originalRess = tfCfg.Module.DataResources
		}

		if len(originalRess) == 0 {
			continue
		}

		// Select resources.
		selectedRess := make(TerraformResources)

		for rn := range originalRess {
			if !slices.Contains(p.ResourceTypes, originalRess[rn].Type) {
				continue
			}

			if len(p.ResourceNames) != 0 &&
				!slices.Contains(p.ResourceNames, originalRess[rn].Name) {
				continue
			}

			selectedRess[rn] = originalRess[rn]
		}

		// Operate.
		operatedRess, err := Operate(selectedRess, &p, cfg.PathSyntax)
		if err != nil {
			return nil, fmt.Errorf("erorr operating on resources: %w", err)
		}

		for rn := range operatedRess {
			originalRess[rn] = operatedRess[rn]
		}
	}

	return tfCfg, nil
}
