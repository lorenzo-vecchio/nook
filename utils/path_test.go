package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpandPathWithTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	assert.NoError(t, err)
	result := ExpandPath("~/projects/nook")
	assert.Equal(t, filepath.Join(home, "projects", "nook"), result)
}

func TestExpandPathWithoutTilde(t *testing.T) {
	result := ExpandPath("/usr/local/bin")
	assert.Equal(t, "/usr/local/bin", result)
}

func TestExpandPathPlain(t *testing.T) {
	result := ExpandPath("relative/path")
	assert.Equal(t, "relative/path", result)
}

func TestResolvePathAbsolute(t *testing.T) {
	result := ResolvePath("/some/base", "/absolute/path")
	assert.Equal(t, filepath.Clean("/absolute/path"), result)
}

func TestResolvePathRelative(t *testing.T) {
	result := ResolvePath("/base/dir", "relative/path")
	assert.Equal(t, filepath.Join("/base/dir", "relative/path"), result)
}

func TestResolvePathWithDotSlash(t *testing.T) {
	result := ResolvePath("/base/dir", "./subdir/file.txt")
	expected := filepath.Clean("/base/dir/subdir/file.txt")
	assert.Equal(t, expected, result)
}

func TestResolvePathWithDotDot(t *testing.T) {
	result := ResolvePath("/base/dir", "../other/file.txt")
	expected := filepath.Clean("/base/other/file.txt")
	assert.Equal(t, expected, result)
}

func TestEnsureDirCreates(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "a", "b", "c")
	err := EnsureDir(testPath)
	assert.NoError(t, err)
	info, err := os.Stat(testPath)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestEnsureDirExistsNoError(t *testing.T) {
	tmpDir := t.TempDir()
	err := EnsureDir(tmpDir)
	assert.NoError(t, err)
}

func TestFileExistsTrue(t *testing.T) {
	tmpDir := t.TempDir()
	f := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(f, []byte("hello"), 0o644)
	assert.NoError(t, err)
	assert.True(t, FileExists(f))
}

func TestFileExistsFalse(t *testing.T) {
	assert.False(t, FileExists("/nonexistent/file.txt"))
}

func TestFileExistsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	assert.False(t, FileExists(tmpDir))
}
