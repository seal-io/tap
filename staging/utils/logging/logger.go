package logging

import (
	"io"
	"log"
	"os"
	"strings"
	"syscall"

	hclog "github.com/hashicorp/go-hclog"
)

const (
	envLog        = "TF_LOG"
	envLogCore    = "TF_LOG_CORE"
	envLogFile    = "TF_LOG_PATH"
	envTmpLogPath = "TF_TEMP_LOG_PATH"
)

var (
	lg       hclog.Logger
	lgWriter io.Writer
)

func init() {
	SetLogger(newLogger(""))
}

// Logger returns the global logger.
func Logger() hclog.Logger {
	return lg
}

// SetLogger sets the global logger.
func SetLogger(l hclog.Logger) {
	lg = l
	lgWriter = lg.StandardWriter(&hclog.StandardLoggerOptions{
		InferLevels: true,
	})

	// Redirect the output of the logger.
	log.SetFlags(0)
	log.SetPrefix("")
	log.SetOutput(lgWriter)
}

func newLogger(name string) hclog.Logger {
	lgOutput := io.Writer(os.Stderr)

	if lgPath := os.Getenv(envLogFile); lgPath != "" {
		f, err := os.OpenFile(lgPath, syscall.O_CREAT|syscall.O_RDWR|syscall.O_APPEND, 0o666)
		if err == nil {
			lgOutput = f
		}
	}

	lvl, json := getLogLevel()
	lg := hclog.NewInterceptLogger(&hclog.LoggerOptions{
		Name:              name,
		Level:             lvl,
		Output:            lgOutput,
		IndependentLevels: true,
		JSONFormat:        json,
	})

	if lgTmpPath := os.Getenv(envTmpLogPath); lgTmpPath != "" {
		f, err := os.OpenFile(lgTmpPath, syscall.O_CREAT|syscall.O_RDWR|syscall.O_APPEND, 0o666)
		if err == nil {
			lg.RegisterSink(hclog.NewSinkAdapter(&hclog.LoggerOptions{
				Level:  hclog.Trace,
				Output: f,
			}))
		}
	}

	return lg
}

func getLogLevel() (hclog.Level, bool) {
	var json bool

	lvl := strings.ToUpper(os.Getenv(envLog))
	if lvl == "" {
		lvl = strings.ToUpper(os.Getenv(envLogCore))
	}

	if lvl == "JSON" {
		json = true
	}

	return parseLogLevel(lvl), json
}

func parseLogLevel(lvl string) hclog.Level {
	switch lvl {
	case "":
		return hclog.Off
	case "JSON":
		lvl = "TRACE"
	}

	for _, l := range []string{
		// The log level names that Terraform recognizes.
		"TRACE",
		"DEBUG",
		"INFO",
		"WARN",
		"ERROR",
		"OFF",
	} {
		if lvl == l {
			return hclog.LevelFromString(lvl)
		}
	}

	return hclog.Trace
}
