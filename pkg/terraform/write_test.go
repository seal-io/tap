package terraform

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrite(t *testing.T) {
	var (
		testCasesLoadDataDir = filepath.Join("testdata", "load")
		testCasesDataDir     = filepath.Join("testdata", "write")
	)

	testCases, err := os.ReadDir(testCasesLoadDataDir)
	if err != nil {
		t.Fatalf("failed to read testdata: %v", err)
	}

	for _, tc := range testCases {
		if !tc.IsDir() {
			continue
		}

		cfg, err := Load(filepath.Join(testCasesLoadDataDir, tc.Name()))
		if !assert.NoErrorf(t, err, "terraform load %s", tc.Name()) {
			continue
		}

		var actualBuff bytes.Buffer

		err = Write(cfg, &actualBuff)
		if !assert.NoErrorf(t, err, "terraform write %s", tc.Name()) {
			continue
		}

		expected, err := os.ReadFile(filepath.Join(testCasesDataDir, tc.Name(), "main.tf"))
		if !assert.NoErrorf(t, err, "expected read %s", tc.Name()) {
			continue
		}

		assert.Equal(t, string(expected), actualBuff.String())
	}
}
