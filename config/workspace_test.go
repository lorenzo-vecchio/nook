package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func validWorkspace() *WorkspaceConfig {
	return &WorkspaceConfig{
		Name:        "my-project",
		Description: "My project workspace",
		Environments: map[string]Environment{
			"dev": {
				EnvFile: ".env.dev",
				Services: []Service{
					{
						Provider: "vscode",
						Folder:   "/home/user/project",
					},
				},
			},
		},
	}
}

func TestLoadWorkspace_Valid(t *testing.T) {
	dir := t.TempDir()
	ws := validWorkspace()
	wsPath := filepath.Join(dir, "workspace.yaml")

	data, err := yaml.Marshal(ws)
	require.NoError(t, err)
	err = os.WriteFile(wsPath, data, 0644)
	require.NoError(t, err)

	loaded, err := LoadWorkspace(wsPath)
	require.NoError(t, err)
	require.NotNil(t, loaded)
	assert.Equal(t, ws.Name, loaded.Name)
	assert.Equal(t, ws.Description, loaded.Description)
	assert.Equal(t, ws.Environments["dev"].Services[0].Provider, loaded.Environments["dev"].Services[0].Provider)
}

func TestLoadWorkspace_FileNotFound(t *testing.T) {
	_, err := LoadWorkspace("/nonexistent/workspace.yaml")
	assert.Error(t, err)
}

func TestLoadWorkspace_InvalidYaml(t *testing.T) {
	dir := t.TempDir()
	wsPath := filepath.Join(dir, "workspace.yaml")
	err := os.WriteFile(wsPath, []byte("invalid: yaml: [broken\n"), 0644)
	require.NoError(t, err)

	_, err = LoadWorkspace(wsPath)
	assert.Error(t, err)
}

func TestSaveWorkspace(t *testing.T) {
	dir := t.TempDir()
	ws := validWorkspace()
	wsPath := filepath.Join(dir, "workspace.yaml")

	err := SaveWorkspace(ws, wsPath)
	require.NoError(t, err)

	_, err = os.Stat(wsPath)
	require.NoError(t, err)

	loaded, err := LoadWorkspace(wsPath)
	require.NoError(t, err)
	assert.Equal(t, ws.Name, loaded.Name)
	assert.Equal(t, ws.Description, loaded.Description)
	assert.Equal(t, ws.Environments, loaded.Environments)
}

func TestSaveAndReloadCycle(t *testing.T) {
	dir := t.TempDir()
	ws := &WorkspaceConfig{
		Name: "test-workspace",
		Environments: map[string]Environment{
			"prod": {
				Services: []Service{
					{
						Provider: "chrome",
						URLs:     []string{"https://example.com"},
					},
					{
						Provider: "docker",
						File:     "docker-compose.yml",
						Profile:  "prod",
					},
				},
			},
			"staging": {
				EnvFile: ".env.staging",
				Services: []Service{
					{
						Provider: "command",
						Cmd:      "echo hello",
						Cwd:      "/tmp",
					},
				},
			},
		},
	}

	wsPath := filepath.Join(dir, "workspace.yaml")
	err := SaveWorkspace(ws, wsPath)
	require.NoError(t, err)

	loaded, err := LoadWorkspace(wsPath)
	require.NoError(t, err)
	require.NotNil(t, loaded)
	assert.Equal(t, ws.Name, loaded.Name)
	assert.Equal(t, len(ws.Environments), len(loaded.Environments))
	assert.Equal(t, ws.Environments["prod"].Services[0].Provider, loaded.Environments["prod"].Services[0].Provider)
	assert.Equal(t, ws.Environments["prod"].Services[0].URLs, loaded.Environments["prod"].Services[0].URLs)
	assert.Equal(t, ws.Environments["prod"].Services[1].Profile, loaded.Environments["prod"].Services[1].Profile)
	assert.Equal(t, ws.Environments["staging"].Services[0].Cmd, loaded.Environments["staging"].Services[0].Cmd)
}

func TestSaveWorkspace_WriteFails(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod not supported on Windows")
	}
	dir := t.TempDir()
	err := os.Chmod(dir, 0555)
	require.NoError(t, err)
	defer func() { _ = os.Chmod(dir, 0755) }()

	ws := validWorkspace()
	wsPath := filepath.Join(dir, "workspace.yaml")
	err = SaveWorkspace(ws, wsPath)
	assert.Error(t, err)
}

func TestWorkspace_NewFieldsRoundTrip(t *testing.T) {
	ws := &WorkspaceConfig{
		Name: "launch-order-test",
		Environments: map[string]Environment{
			"dev": {
				EnvFile: ".env.dev",
				WaitForComposeHealthy: true,
				Services: []Service{
					{
						Provider: "vscode",
						Folder:   "/project",
						Order:    1,
						DelayMs:  500,
					},
					{
						Provider:   "docker",
						File:       "docker-compose.yml",
						Order:     2,
						DelayMs:   0,
						ReadyCheck: &ReadyCheck{Cmd: "curl http://localhost:8080/health", IntervalMs: 1000, TimeoutMs: 30000},
					},
				},
			},
		},
	}

	dir := t.TempDir()
	wsPath := filepath.Join(dir, "workspace.yaml")

	err := SaveWorkspace(ws, wsPath)
	require.NoError(t, err)

	loaded, err := LoadWorkspace(wsPath)
	require.NoError(t, err)
	require.NotNil(t, loaded)

	dev := loaded.Environments["dev"]
	assert.True(t, dev.WaitForComposeHealthy)

	svc0 := dev.Services[0]
	assert.Equal(t, 1, svc0.Order)
	assert.Equal(t, 500, svc0.DelayMs)
	assert.Nil(t, svc0.ReadyCheck)

	svc1 := dev.Services[1]
	assert.Equal(t, 2, svc1.Order)
	assert.Equal(t, 0, svc1.DelayMs)
	require.NotNil(t, svc1.ReadyCheck)
	assert.Equal(t, "curl http://localhost:8080/health", svc1.ReadyCheck.Cmd)
	assert.Equal(t, 1000, svc1.ReadyCheck.IntervalMs)
	assert.Equal(t, 30000, svc1.ReadyCheck.TimeoutMs)
}

func TestWorkspace_NewFieldsOmitempty(t *testing.T) {
	ws := &WorkspaceConfig{
		Name: "omitempty-test",
		Environments: map[string]Environment{
			"dev": {
				Services: []Service{
					{Provider: "chrome", Folder: "/p"},
				},
			},
		},
	}

	data, err := yaml.Marshal(ws)
	require.NoError(t, err)

	yamlStr := string(data)
	assert.NotContains(t, yamlStr, "delay_ms")
	assert.NotContains(t, yamlStr, "order")
	assert.NotContains(t, yamlStr, "ready_check")
	assert.NotContains(t, yamlStr, "wait_for_compose_healthy")
}

func TestWorkspace_WithTerminals(t *testing.T) {
	ws := &WorkspaceConfig{
		Name: "term-test",
		Environments: map[string]Environment{
			"dev": {
				Services: []Service{
					{
						Provider: "vscode",
						Folder:   "/project",
						Terminals: []Terminal{
							{Name: "Server", Directory: "/project/server", Command: "npm run dev"},
							{Name: "Client", Directory: "/project/client", Command: "npm start"},
						},
					},
				},
			},
		},
	}

	dir := t.TempDir()
	wsPath := filepath.Join(dir, "workspace.yaml")
	err := SaveWorkspace(ws, wsPath)
	require.NoError(t, err)

	loaded, err := LoadWorkspace(wsPath)
	require.NoError(t, err)
	assert.Len(t, loaded.Environments["dev"].Services[0].Terminals, 2)
	assert.Equal(t, "Server", loaded.Environments["dev"].Services[0].Terminals[0].Name)
	assert.Equal(t, "npm run dev", loaded.Environments["dev"].Services[0].Terminals[0].Command)
}
