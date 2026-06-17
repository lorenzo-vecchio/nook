package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/anomalyco/nook/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListCommand_Use(t *testing.T) {
	cmd := NewListCmd()
	assert.Equal(t, "list", cmd.Use)
}

func TestListCommand_Help(t *testing.T) {
	cmd := NewListCmd()
	assert.NotEmpty(t, cmd.Short)
}

func TestListCommand_NoWorkspaces(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cmd := NewListCmd()
	output := captureStdout(t, func() {
		err := cmd.Execute()
		assert.NoError(t, err)
	})
	assert.Contains(t, output, "No workspaces found")
}

func TestListCommand_WithWorkspaces(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	scanPath := filepath.Join(dir, "workspaces")
	err := os.MkdirAll(scanPath, 0755)
	require.NoError(t, err)

	cfg := &config.GlobalConfig{
		ScanPaths: []string{scanPath},
	}
	err = config.SaveGlobalConfig(cfg)
	require.NoError(t, err)

	wsDir := filepath.Join(scanPath, "my-project")
	err = os.MkdirAll(wsDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(wsDir, "workspace.yaml"), []byte("name: my-project\ndescription: My project\nenvironments:\n  dev:\n    services:\n      - provider: command\n        cmd: echo hello\n"), 0644)
	require.NoError(t, err)

	cmd := NewListCmd()
	output := captureStdout(t, func() {
		err := cmd.Execute()
		assert.NoError(t, err)
	})
	assert.Contains(t, output, "my-project")
	assert.Contains(t, output, "My project")
	assert.Contains(t, output, scanPath)
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	out := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		assert.NoError(t, err)
		out <- buf.String()
	}()

	fn()

	w.Close()
	os.Stdout = old
	return <-out
}
