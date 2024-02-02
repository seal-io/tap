package workingdir

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/seal-io/tap/pkg/tap"
	"github.com/seal-io/tap/pkg/terraform"
)

// Setup prepares the working directory.
func Setup(args []string) ([]string, error) {
	// Get root dir.
	rootDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Get working dir.
	workingDir, args, err := extractChdirOption(args)
	if err != nil {
		return nil, err
	}

	if workingDir == "" {
		workingDir = rootDir
	}

	// Load tap config.
	cfg, err := tap.Load(workingDir)
	if err != nil {
		return nil, fmt.Errorf("error loading tap configuration: %w", err)
	}

	if cfg != nil {
		// Clean tap.
		tapDir := filepath.Join(workingDir, ".tap")
		if err = cleanDir(tapDir); err != nil {
			return nil, fmt.Errorf("error preparing the tap directory")
		}

		// Load terraform config.
		tfcfg, err := terraform.Load(workingDir)
		if err != nil {
			return nil, fmt.Errorf("error loading terraform configuration: %w", err)
		}

		// Apply tap configuration.
		tfcfg, err = tap.Apply(tfcfg, cfg)
		if err != nil {
			return nil, fmt.Errorf("error applying tap configuration: %w", err)
		}

		// Copy the working dir to the tap dir.
		if err = copyDir(workingDir, tapDir); err != nil {
			return nil, fmt.Errorf("error copying the working directory")
		}

		// Write the terraform configuration.
		f, err := os.OpenFile(filepath.Join(tapDir, "main.tf"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			return nil, fmt.Errorf("error creating terraform configuration: %w", err)
		}

		defer func() { _ = f.Close() }()

		if err = terraform.Write(tfcfg, f); err != nil {
			return nil, fmt.Errorf("error writing terraform configuration: %w", err)
		}

		// Mutate the working dir if tap is configured.
		workingDir = tapDir
	}

	// Create new arguments.
	newArgs := make([]string, len(args)+1)
	newArgs[0] = "-chdir=" + workingDir
	copy(newArgs[1:], args)

	return newArgs, nil
}
