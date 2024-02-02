package main

import (
	"errors"
	"log"
	"os"
	"os/exec"

	"golang.org/x/exp/slices"

	"github.com/seal-io/tap/cmd"
	"github.com/seal-io/tap/utils/signals"
)

func main() {
	var (
		args   = os.Args[:1]
		tfBins = []string{"terraform", "tofu"}
	)

	if len(args) != 0 && slices.Contains(tfBins, args[0]) {
		args = args[1:]
		tfBins = []string{args[0]}
	}

	var (
		tfBin string
		err   error
	)

	for _, b := range tfBins {
		tfBin, err = exec.LookPath(b)
		if err == nil {
			break
		}
	}

	if err != nil {
		log.Printf("[ERROR] Could not load terraform executable binary\n")
		os.Exit(1)
	}

	err = cmd.Delegate(signals.SetupSignalHandler(), tfBin, os.Args[1:])
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}

		os.Exit(1)
	}
}
