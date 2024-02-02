package tap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	testCasesDataDir := filepath.Join("testdata", "load")

	testCases, err := os.ReadDir(testCasesDataDir)
	if err != nil {
		t.Fatalf("failed to read testdata: %v", err)
	}

	for _, tc := range testCases {
		if !tc.IsDir() {
			continue
		}

		cfg, err := Load(filepath.Join(testCasesDataDir, tc.Name()))

		switch n := tc.Name(); {
		case n == "none":
			assert.NoError(t, err, n)
		case strings.HasPrefix(n, "invalid_"):
			assert.Error(t, err, n)
		default:
			assert.NoError(t, err, n)
			assert.NotNil(t, cfg, n)
			assert.Greater(t, len(cfg.Patches), 0, n)
		}
	}
}
