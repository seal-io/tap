package terraform

import (
	"os"
	"path/filepath"
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

		_, err := Load(filepath.Join(testCasesDataDir, tc.Name()))
		assert.NoError(t, err)
	}
}
