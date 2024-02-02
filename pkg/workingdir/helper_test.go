package workingdir

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func Test_copyDir_cleanDir(t *testing.T) {
	var err error

	src := filepath.Join("testdata", "copy_clean_dir")
	dst := filepath.Join(src, ".tap")

	beforeClean, err := os.ReadFile(filepath.Join(dst, "terraform.tfstate"))
	if err != nil {
		t.Fatalf("failed to read terraform.tfstate before cleaning: %v", err)
	}

	err = cleanDir(filepath.Join(src, ".tap"))
	assert.NoError(t, err, "failed to clean directory")

	afterClean, err := os.ReadFile(filepath.Join(dst, "terraform.tfstate"))
	if err != nil {
		t.Fatalf("failed to read terraform.tfstate after cleaning: %v", err)
	}

	assert.Equalf(t, beforeClean, afterClean, "terraform.tfstate file content changed after cleaning")

	err = copyDir(src, dst)
	assert.NoError(t, err, "failed to copy directory")

	afterCopy, err := os.ReadFile(filepath.Join(dst, "terraform.tfstate"))
	if err != nil {
		t.Fatalf("failed to read terraform.tfstate after copying: %v", err)
	}

	assert.Equalf(t, beforeClean, afterCopy, "terraform.tfstate file content changed after cleaning")
	{
		dstFs := afero.Afero{Fs: afero.NewBasePathFs(afero.NewOsFs(), dst)}

		found, err := afero.Glob(dstFs, "*.tf")
		if err != nil {
			t.Fatalf("failed to glob terraform files: %v", err)
		}

		assert.Len(t, found, 0, "unexpected number of terraform files")
	}

	err = cleanDir(filepath.Join(src, ".tap"))
	assert.NoError(t, err, "failed to clean directory")
}
