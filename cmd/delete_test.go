package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/anomalyco/nook/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPrompter struct {
	selectFn   func(label string, options []string, defaultOption string) (string, error)
	confirmFn  func(label string, defaultVal bool) (bool, error)
}

func (m *mockPrompter) Select(label string, options []string, defaultOption string) (string, error) {
	if m.selectFn != nil {
		return m.selectFn(label, options, defaultOption)
	}
	return options[0], nil
}

func (m *mockPrompter) MultiSelect(label string, options []string, defaults []string) ([]string, error) {
	return nil, nil
}

func (m *mockPrompter) Input(label string, defaultVal string) (string, error) {
	return "", nil
}

func (m *mockPrompter) Confirm(label string, defaultVal bool) (bool, error) {
	if m.confirmFn != nil {
		return m.confirmFn(label, defaultVal)
	}
	return true, nil
}

func TestDeleteCommand_Use(t *testing.T) {
	cmd := NewDeleteCmd()
	assert.Equal(t, "delete [name]", cmd.Use)
}

func TestDeleteCommand_Help(t *testing.T) {
	cmd := NewDeleteCmd()
	assert.NotEmpty(t, cmd.Short)
}

func TestDeleteCommand_InvalidName(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := &config.GlobalConfig{
		ScanPaths: []string{dir},
	}
	err := config.SaveGlobalConfig(cfg)
	require.NoError(t, err)

	oldPrompter := deletePrompter
	deletePrompter = &mockPrompter{}
	defer func() { deletePrompter = oldPrompter }()

	cmd := NewDeleteCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"nonexistent"})
	err = cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteCommand_HandCreatedWorkspace(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

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

	oldPrompter := deletePrompter
	deletePrompter = &mockPrompter{
		confirmFn: func(label string, defaultVal bool) (bool, error) {
			return true, nil
		},
	}
	defer func() { deletePrompter = oldPrompter }()

	cmd := NewDeleteCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"my-workspace"})
	err = cmd.Execute()
	assert.NoError(t, err)

	_, err = os.Stat(wsDir)
	assert.True(t, os.IsNotExist(err))
}

func TestDeleteCommand_ConfirmDeclined(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

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

	oldPrompter := deletePrompter
	deletePrompter = &mockPrompter{
		confirmFn: func(label string, defaultVal bool) (bool, error) {
			return false, nil
		},
	}
	defer func() { deletePrompter = oldPrompter }()

	cmd := NewDeleteCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"my-workspace"})
	err = cmd.Execute()
	assert.NoError(t, err)

	_, err = os.Stat(wsDir)
	assert.NoError(t, err)
}
