package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lorenzo-vecchio/nook/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceTypeToProvider(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"VS Code", "vscode"},
		{"DBeaver", "dbeaver"},
		{"Chrome", "chrome"},
		{"Docker Compose", "docker"},
		{"Custom Command", "command"},
		{"Unknown", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, serviceTypeToProvider(tt.input))
		})
	}
}

func TestInitCmd_CreatesWorkspace(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	inputCalls := 0
	confirmCalls := 0

	mp := &mockPrompter{
		inputFn: func(label, defaultVal string) (string, error) {
			inputCalls++
			vals := []string{
				"test-workspace",
				"Test description",
				"dev",
				"",
				"/home/user/project",
				"Server",
				"/home/user/project/server",
				"npm run dev",
				"https://example.com,https://test.com",
			}
			if inputCalls-1 < len(vals) {
				return vals[inputCalls-1], nil
			}
			return "", nil
		},
		confirmFn: func(label string, defaultVal bool) (bool, error) {
			confirmCalls++
			vals := []bool{true, false, false}
			if confirmCalls-1 < len(vals) {
				return vals[confirmCalls-1], nil
			}
			return false, nil
		},
		multiSelectFn: func(label string, options, defaults []string) ([]string, error) {
			return []string{"VS Code", "Chrome"}, nil
		},
	}

	cmd := NewInitCmd(mp)
	err := cmd.Execute()
	require.NoError(t, err)

	wsPath := filepath.Join(homeDir, ".nook", "workspaces", "test-workspace", "workspace.yaml")
	_, err = os.Stat(wsPath)
	require.NoError(t, err)

	ws, err := config.LoadWorkspace(wsPath)
	require.NoError(t, err)
	require.NotNil(t, ws)
	assert.Equal(t, "test-workspace", ws.Name)
	assert.Equal(t, "Test description", ws.Description)
	assert.Contains(t, ws.Environments, "dev")

	services := ws.Environments["dev"].Services
	require.Len(t, services, 2)

	assert.Equal(t, "vscode", services[0].Provider)
	assert.Equal(t, "/home/user/project", services[0].Folder)
	require.Len(t, services[0].Terminals, 1)
	assert.Equal(t, "Server", services[0].Terminals[0].Name)
	assert.Equal(t, "/home/user/project/server", services[0].Terminals[0].Directory)
	assert.Equal(t, "npm run dev", services[0].Terminals[0].Command)

	assert.Equal(t, "chrome", services[1].Provider)
	require.Len(t, services[1].URLs, 2)
	assert.Equal(t, "https://example.com", services[1].URLs[0])
	assert.Equal(t, "https://test.com", services[1].URLs[1])

	dotDir := filepath.Join(homeDir, ".nook", "workspaces", "test-workspace", ".workspace")
	info, err := os.Stat(dotDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestInitCmd_CreatesWorkspace_WithDBeaverDockerCommand(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	inputCalls := 0
	confirmCalls := 0

	mp := &mockPrompter{
		inputFn: func(label, defaultVal string) (string, error) {
			inputCalls++
			vals := []string{
				"multi-svc",
				"Multiple services",
				"dev",
				"",
				"connection-string",
				"docker-compose.yml",
				"prod",
				"echo hello",
				"/tmp/work",
			}
			if inputCalls-1 < len(vals) {
				return vals[inputCalls-1], nil
			}
			return "", nil
		},
		confirmFn: func(label string, defaultVal bool) (bool, error) {
			confirmCalls++
			vals := []bool{false}
			if confirmCalls-1 < len(vals) {
				return vals[confirmCalls-1], nil
			}
			return false, nil
		},
		multiSelectFn: func(label string, options, defaults []string) ([]string, error) {
			return []string{"DBeaver", "Docker Compose", "Custom Command"}, nil
		},
	}

	cmd := NewInitCmd(mp)
	err := cmd.Execute()
	require.NoError(t, err)

	wsPath := filepath.Join(homeDir, ".nook", "workspaces", "multi-svc", "workspace.yaml")
	_, err = os.Stat(wsPath)
	require.NoError(t, err)

	ws, err := config.LoadWorkspace(wsPath)
	require.NoError(t, err)
	require.NotNil(t, ws)
	assert.Equal(t, "multi-svc", ws.Name)

	services := ws.Environments["dev"].Services
	require.Len(t, services, 3)

	assert.Equal(t, "dbeaver", services[0].Provider)
	assert.Equal(t, "connection-string", services[0].Connection)

	assert.Equal(t, "docker", services[1].Provider)
	assert.Equal(t, "docker-compose.yml", services[1].File)
	assert.Equal(t, "prod", services[1].Profile)

	assert.Equal(t, "command", services[2].Provider)
	assert.Equal(t, "echo hello", services[2].Cmd)
	assert.Equal(t, "/tmp/work", services[2].Cwd)
}

func TestInitCmd_ValidationFailure(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	inputCalls := 0
	confirmCalls := 0

	mp := &mockPrompter{
		inputFn: func(label, defaultVal string) (string, error) {
			inputCalls++
			vals := []string{"", "", "dev", "", "/home/user/project"}
			if inputCalls-1 < len(vals) {
				return vals[inputCalls-1], nil
			}
			return "", nil
		},
		confirmFn: func(label string, defaultVal bool) (bool, error) {
			confirmCalls++
			vals := []bool{true, false, false}
			if confirmCalls-1 < len(vals) {
				return vals[confirmCalls-1], nil
			}
			return false, nil
		},
		multiSelectFn: func(label string, options, defaults []string) ([]string, error) {
			return []string{"VS Code"}, nil
		},
	}

	cmd := NewInitCmd(mp)
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}
