package upload

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateFile(t *testing.T) {
	const filePerm = 0o600

	dir := t.TempDir()
	path := filepath.Join(dir, "note.txt")

	assert.NoError(t, os.WriteFile(path, []byte("hello"), filePerm))
	assert.NoError(t, validateFile(path))

	assert.Error(t, validateFile(dir))
	assert.Error(t, validateFile(filepath.Join(dir, "missing.txt")))
}
