package tap

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/seal-io/tap/pkg/terraform"
)

func TestApply(t *testing.T) {
	testCasesDataDir := filepath.Join("testdata", "apply")

	testCases, err := os.ReadDir(testCasesDataDir)
	if err != nil {
		t.Fatalf("failed to read testdata: %v", err)
	}

	for _, tc := range testCases {
		if !tc.IsDir() {
			continue
		}

		tfCfg, err := terraform.Load(filepath.Join(testCasesDataDir, tc.Name()))
		if !assert.NoErrorf(t, err, "terraform load %s", tc.Name()) {
			continue
		}

		cfg, err := Load(filepath.Join(testCasesDataDir, tc.Name()))
		if !assert.NoErrorf(t, err, "tap load %s", tc.Name()) {
			continue
		}

		tfCfg, err = Apply(tfCfg, cfg)
		if !assert.NoErrorf(t, err, "error appling %s", tc.Name()) {
			continue
		}

		var actualBuff bytes.Buffer

		err = terraform.Write(tfCfg, &actualBuff)
		if !assert.NoErrorf(t, err, "terraform write %s", tc.Name()) {
			continue
		}

		expectedBytes, err := os.ReadFile(filepath.Join(testCasesDataDir, tc.Name(), "expected", "main.tf"))
		if !assert.NoErrorf(t, err, "error reading expected %s", tc.Name()) {
			continue
		}

		assert.Equal(t, string(expectedBytes), actualBuff.String(), "apply %s", tc.Name())
	}
}
