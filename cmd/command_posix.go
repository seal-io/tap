//go:build !windows

package cmd

import (
	"context"
	"os"
	"os/exec"
	"syscall"
)

func newCommandContext(ctx context.Context, cmd string, args []string) *exec.Cmd {
	ce := exec.CommandContext(ctx, cmd, args...)
	ce.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	ce.Stdin, ce.Stdout, ce.Stderr = os.Stdin, os.Stdout, os.Stderr

	return ce
}
