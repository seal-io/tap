package cmd

import (
	"context"
	"os"
	"os/exec"
)

func newCommandContext(ctx context.Context, cmd string, args []string) *exec.Cmd {
	ce := exec.CommandContext(ctx, cmd, args...)
	ce.Stdin, ce.Stdout, ce.Stderr = os.Stdin, os.Stdout, os.Stderr

	return ce
}
