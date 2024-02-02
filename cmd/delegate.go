package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/seal-io/tap/pkg/workingdir"
	"github.com/seal-io/tap/utils/set"
	"github.com/seal-io/tap/utils/version"

	_ "github.com/seal-io/tap/utils/logging"
)

func Delegate(ctx context.Context, cmd string, args []string) (err error) {
	// Setup working dir.
	args, err = workingdir.Setup(args)
	if err != nil {
		return fmt.Errorf("error setting up the working directory: %w", err)
	}

	// Delegate.
	return delegate(ctx, cmd, args)
}

func delegate(ctx context.Context, cmd string, args []string) error {
	// Construct execution.
	ce := newCommandContext(ctx, cmd, args)

	// Hijack version command.
	as := set.New(args...)
	if as.HasAny("version", "-v", "-version", "--version") {
		if !as.Has("-json") {
			fmt.Printf("Tap %s\n", version.Get())
		} else {
			// Merge the version information into the JSON output.
			var buf bytes.Buffer
			ce.Stdout = &buf
			ce.Stderr = &buf

			if err := ce.Run(); err == nil {
				var ot map[string]any
				_ = json.Unmarshal(buf.Bytes(), &ot)
				ot["tap_version"] = version.Version
				ot["tap_git_commit"] = version.GitCommit
				bs, _ := json.MarshalIndent(ot, "", "  ")
				_, _ = fmt.Fprintf(os.Stdout, "%s\n", string(bs))

				return nil
			}
		}
	}

	return ce.Run()
}
