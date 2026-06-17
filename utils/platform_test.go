package utils

import (
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsMacOS(t *testing.T) {
	assert.Equal(t, runtime.GOOS == "darwin", IsMacOS())
}

func TestIsLinux(t *testing.T) {
	assert.Equal(t, runtime.GOOS == "linux", IsLinux())
}

func TestIsWindows(t *testing.T) {
	assert.Equal(t, runtime.GOOS == "windows", IsWindows())
}

func TestDefaultEditorRespectsEnv(t *testing.T) {
	prev := os.Getenv("EDITOR")
	os.Setenv("EDITOR", "nano")
	defer os.Setenv("EDITOR", prev)
	assert.Equal(t, "nano", DefaultEditor())
}

func TestDefaultEditorFallback(t *testing.T) {
	prev := os.Getenv("EDITOR")
	os.Unsetenv("EDITOR")
	defer os.Setenv("EDITOR", prev)
	expected := "vim"
	if runtime.GOOS == "windows" {
		expected = "notepad"
	}
	assert.Equal(t, expected, DefaultEditor())
}

func TestShellCommand(t *testing.T) {
	shell, args := ShellCommand("echo hello")
	if runtime.GOOS == "windows" {
		assert.Equal(t, "cmd", shell)
		assert.Equal(t, []string{"/c", "echo hello"}, args)
	} else {
		assert.Equal(t, "sh", shell)
		assert.Equal(t, []string{"-c", "echo hello"}, args)
	}
}
