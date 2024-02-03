// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !solaris
// +build !solaris

// The readline library we use doesn't currently support solaris so
// we just build tag it off.

package command

import (
	"fmt"
	"io"
	"os"

	"github.com/hashicorp/terraform/internal/repl"

	"github.com/chzyer/readline"
	"github.com/mitchellh/cli"
)

func (c *ConsoleCommand) modeInteractive(session *repl.Session, ui cli.Ui) int {
	// Configure input
	l, err := readline.NewEx(&readline.Config{
		Prompt:            "> ",
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
		Stdin:             os.Stdin,
		Stdout:            os.Stdout,
		Stderr:            os.Stderr,
	})
	if err != nil {
		c.Ui.Error(fmt.Sprintf(
			"Error initializing console: %s",
			err))
		return 1
	}
	defer l.Close()

	for {
		// Read a line
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		out, exit, diags := session.Handle(line)
		if diags.HasErrors() {
			c.showDiagnostics(diags)
		}
		if exit {
			break
		}

		ui.Output(out)
	}

	return 0
}
