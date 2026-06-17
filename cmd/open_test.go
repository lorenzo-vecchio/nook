package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/anomalyco/nook/config"
	"github.com/anomalyco/nook/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockProvider struct {
	name         string
	launchCalled bool
	detectResult bool
	launchErr    error
}

func (m *mockProvider) Name() string { return m.name }

func (m *mockProvider) Detect() (bool, error) { return m.detectResult, nil }

func (m *mockProvider) Launch(_ context.Context, _ config.Service, _ string, _ map[string]string) error {
	m.launchCalled = true
	return m.launchErr
}

func createTestWorkspace(t *testing.T, dir, name string, envs map[string]config.Environment) {
	ws := &config.WorkspaceConfig{
		Name:         name,
		Environments: envs,
	}
	err := config.SaveWorkspace(ws, filepath.Join(dir, "workspace.yaml"))
	require.NoError(t, err)
}

func TestOpenCmd_ByNameWithEnvFlag(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("XDG_CONFIG_HOME", homeDir)

	wsDir := filepath.Join(homeDir, ".nook", "workspaces", "test-ws")
	err := os.MkdirAll(wsDir, 0755)
	require.NoError(t, err)

	createTestWorkspace(t, wsDir, "test-ws", map[string]config.Environment{
		"dev": {
			Services: []config.Service{
				{Provider: "mock", Folder: "/project"},
			},
		},
	})

	cfg := &config.GlobalConfig{ScanPaths: []string{wsDir}}
	require.NoError(t, config.SaveGlobalConfig(cfg))

	mockProv := &mockProvider{name: "mock"}
	provider.Register(mockProv)

	mp := &mockPrompter{
		selectFn: func(label string, options []string, defaultOption string) (string, error) {
			t.Fatal("unexpected Select call")
			return "", nil
		},
		inputFn: func(label, defaultVal string) (string, error) {
			return "", nil
		},
		confirmFn: func(label string, defaultVal bool) (bool, error) {
			return true, nil
		},
		multiSelectFn: func(label string, options, defaults []string) ([]string, error) {
			return nil, nil
		},
	}

	cmd := NewOpenCmd(mp)
	cmd.SetArgs([]string{"test-ws", "--env", "dev"})
	err = cmd.Execute()
	require.NoError(t, err)
	assert.True(t, mockProv.launchCalled)
}

func TestOpenCmd_ByNameNotFound(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("XDG_CONFIG_HOME", homeDir)

	cfg := &config.GlobalConfig{ScanPaths: []string{homeDir}}
	require.NoError(t, config.SaveGlobalConfig(cfg))

	mp := &mockPrompter{
		selectFn: func(label string, options []string, defaultOption string) (string, error) {
			t.Fatal("unexpected Select call")
			return "", nil
		},
		inputFn: func(label, defaultVal string) (string, error) {
			return "", nil
		},
		confirmFn: func(label string, defaultVal bool) (bool, error) {
			return true, nil
		},
		multiSelectFn: func(label string, options, defaults []string) ([]string, error) {
			return nil, nil
		},
	}

	cmd := NewOpenCmd(mp)
	cmd.SetArgs([]string{"nonexistent"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestOpenCmd_NoWorkspaces(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("XDG_CONFIG_HOME", homeDir)

	cfg := &config.GlobalConfig{ScanPaths: []string{homeDir}}
	require.NoError(t, config.SaveGlobalConfig(cfg))

	mp := &mockPrompter{
		selectFn: func(label string, options []string, defaultOption string) (string, error) {
			t.Fatal("unexpected Select call")
			return "", nil
		},
		inputFn: func(label, defaultVal string) (string, error) {
			return "", nil
		},
		confirmFn: func(label string, defaultVal bool) (bool, error) {
			return true, nil
		},
		multiSelectFn: func(label string, options, defaults []string) ([]string, error) {
			return nil, nil
		},
	}

	cmd := NewOpenCmd(mp)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no workspaces found")
}

func TestOpenCmd_SelectWorkspace(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("XDG_CONFIG_HOME", homeDir)

	workspacesDir := filepath.Join(homeDir, ".nook", "workspaces")
	ws1Dir := filepath.Join(workspacesDir, "ws1")
	ws2Dir := filepath.Join(workspacesDir, "ws2")
	err := os.MkdirAll(ws1Dir, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(ws2Dir, 0755)
	require.NoError(t, err)

	createTestWorkspace(t, ws1Dir, "ws1", map[string]config.Environment{
		"dev": {Services: []config.Service{{Provider: "mock", Folder: "/proj1"}}},
	})
	createTestWorkspace(t, ws2Dir, "ws2", map[string]config.Environment{
		"prod": {Services: []config.Service{{Provider: "mock", Folder: "/proj2"}}},
	})

	cfg := &config.GlobalConfig{ScanPaths: []string{workspacesDir}}
	require.NoError(t, config.SaveGlobalConfig(cfg))

	mockProv := &mockProvider{name: "mock"}
	provider.Register(mockProv)

	mp := &mockPrompter{
		selectFn: func(label string, options []string, defaultOption string) (string, error) {
			return "ws1", nil
		},
		inputFn: func(label, defaultVal string) (string, error) {
			return "", nil
		},
		confirmFn: func(label string, defaultVal bool) (bool, error) {
			return true, nil
		},
		multiSelectFn: func(label string, options, defaults []string) ([]string, error) {
			return nil, nil
		},
	}

	cmd := NewOpenCmd(mp)
	cmd.SetArgs([]string{})
	err = cmd.Execute()
	require.NoError(t, err)
	assert.True(t, mockProv.launchCalled)
}

func TestOpenCmd_ProviderNotFound(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("XDG_CONFIG_HOME", homeDir)

	wsDir := filepath.Join(homeDir, ".nook", "workspaces", "bad-provider")
	err := os.MkdirAll(wsDir, 0755)
	require.NoError(t, err)

	createTestWorkspace(t, wsDir, "bad-provider", map[string]config.Environment{
		"dev": {
			Services: []config.Service{
				{Provider: "nonexistent-provider", Folder: "/project"},
			},
		},
	})

	cfg := &config.GlobalConfig{ScanPaths: []string{wsDir}}
	require.NoError(t, config.SaveGlobalConfig(cfg))

	mp := &mockPrompter{
		selectFn: func(label string, options []string, defaultOption string) (string, error) {
			t.Fatal("unexpected Select call")
			return "", nil
		},
		inputFn: func(label, defaultVal string) (string, error) {
			return "", nil
		},
		confirmFn: func(label string, defaultVal bool) (bool, error) {
			return true, nil
		},
		multiSelectFn: func(label string, options, defaults []string) ([]string, error) {
			return nil, nil
		},
	}

	cmd := NewOpenCmd(mp)
	cmd.SetArgs([]string{"bad-provider", "--env", "dev"})
	err = cmd.Execute()
	require.NoError(t, err)
}

func TestOpenCmd_SelectEnvironment(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("XDG_CONFIG_HOME", homeDir)

	wsDir := filepath.Join(homeDir, ".nook", "workspaces", "multi-env")
	err := os.MkdirAll(wsDir, 0755)
	require.NoError(t, err)

	createTestWorkspace(t, wsDir, "multi-env", map[string]config.Environment{
		"dev":  {Services: []config.Service{{Provider: "mock", Folder: "/dev"}}},
		"prod": {Services: []config.Service{{Provider: "mock", Folder: "/prod"}}},
	})

	cfg := &config.GlobalConfig{ScanPaths: []string{wsDir}}
	require.NoError(t, config.SaveGlobalConfig(cfg))

	mockProv := &mockProvider{name: "mock"}
	provider.Register(mockProv)

	selectCalls := 0
	mp := &mockPrompter{
		selectFn: func(label string, options []string, defaultOption string) (string, error) {
			selectCalls++
			if selectCalls == 1 {
				return "prod", nil
			}
			return "", nil
		},
		inputFn: func(label, defaultVal string) (string, error) {
			return "", nil
		},
		confirmFn: func(label string, defaultVal bool) (bool, error) {
			return true, nil
		},
		multiSelectFn: func(label string, options, defaults []string) ([]string, error) {
			return nil, nil
		},
	}

	cmd := NewOpenCmd(mp)
	cmd.SetArgs([]string{"multi-env"})
	err = cmd.Execute()
	require.NoError(t, err)
	assert.True(t, mockProv.launchCalled)
}
