package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/lorenzo-vecchio/nook/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEditCommand_Use(t *testing.T) {
	cmd := NewEditCmd()
	assert.Equal(t, "edit [name]", cmd.Use)
}

func TestEditCommand_Help(t *testing.T) {
	cmd := NewEditCmd()
	assert.NotEmpty(t, cmd.Short)
}

func TestEditCommand_InvalidName(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := &config.GlobalConfig{
		ScanPaths: []string{dir},
	}
	err := config.SaveGlobalConfig(cfg)
	require.NoError(t, err)

	cmd := NewEditCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"nonexistent"})
	err = cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestEditCommand_ValidName(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("EDITOR", "true")

	wsDir := filepath.Join(dir, "my-workspace")
	err := os.MkdirAll(wsDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(wsDir, "workspace.yaml"), []byte("name: my-workspace\ndescription: Test\nenvironments:\n  dev:\n    services:\n      - provider: command\n        cmd: echo hi\n"), 0644)
	require.NoError(t, err)

	cfg := &config.GlobalConfig{
		ScanPaths: []string{dir},
	}
	err = config.SaveGlobalConfig(cfg)
	require.NoError(t, err)

	cmd := NewEditCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"my-workspace"})
	err = cmd.Execute()
	assert.NoError(t, err)
}

func TestEditCommand_WithFlags(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	wsDir := filepath.Join(dir, "my-workspace")
	err := os.MkdirAll(wsDir, 0755)
	require.NoError(t, err)
	yamlContent := "name: my-workspace\ndescription: old desc\nenvironments:\n  dev:\n    services:\n      - provider: command\n        cmd: echo hi\n"
	wsPath := filepath.Join(wsDir, "workspace.yaml")
	err = os.WriteFile(wsPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	cfg := &config.GlobalConfig{
		ScanPaths: []string{dir},
	}
	err = config.SaveGlobalConfig(cfg)
	require.NoError(t, err)

	cmd := NewEditCmd()
	cmd.SetArgs([]string{"my-workspace", "--name", "renamed", "--description", "new desc"})
	err = cmd.Execute()
	require.NoError(t, err)

	ws, err := config.LoadWorkspace(wsPath)
	require.NoError(t, err)
	assert.Equal(t, "renamed", ws.Name)
	assert.Equal(t, "new desc", ws.Description)
}
