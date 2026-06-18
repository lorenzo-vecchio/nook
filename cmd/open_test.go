package cmd

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/lorenzo-vecchio/nook/config"
	"github.com/lorenzo-vecchio/nook/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	launchMu     sync.Mutex
	launchRecord []string
)

type recordProvider struct {
	name string
}

func (p *recordProvider) Name() string { return p.name }

func (p *recordProvider) Detect() (bool, error) { return true, nil }

func (p *recordProvider) Launch(_ context.Context, _ config.Service, _ string, _ map[string]string) error {
	launchMu.Lock()
	launchRecord = append(launchRecord, p.name)
	launchMu.Unlock()
	return nil
}

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

func TestOpenCmd_WithOrdering(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("XDG_CONFIG_HOME", homeDir)

	wsDir := filepath.Join(homeDir, ".nook", "workspaces", "ordered")
	err := os.MkdirAll(wsDir, 0755)
	require.NoError(t, err)

	createTestWorkspace(t, wsDir, "ordered", map[string]config.Environment{
		"dev": {
			Services: []config.Service{
				{Provider: "docker", File: "docker-compose.yml", Order: 1},
				{Provider: "vscode", Folder: "/project", Order: 2, DelayMs: 500},
				{Provider: "chrome", URLs: []string{"http://example.com"}, Order: 0},
			},
		},
	})

	cfg := &config.GlobalConfig{ScanPaths: []string{wsDir}}
	require.NoError(t, config.SaveGlobalConfig(cfg))

	launchRecord = nil

	savedExec := execCommandContext
	execCommandContext = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		launchMu.Lock()
		launchRecord = append(launchRecord, "docker")
		launchMu.Unlock()
		return exec.CommandContext(ctx, "/bin/echo", arg...)
	}
	t.Cleanup(func() { execCommandContext = savedExec })

	provider.Register(&recordProvider{name: "docker"})
	provider.Register(&recordProvider{name: "vscode"})
	provider.Register(&recordProvider{name: "chrome"})

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
	cmd.SetArgs([]string{"ordered", "--env", "dev"})
	err = cmd.Execute()
	require.NoError(t, err)

	launchMu.Lock()
	order := make([]string, len(launchRecord))
	copy(order, launchRecord)
	launchMu.Unlock()

	require.Len(t, order, 3)
	assert.Equal(t, "docker", order[0])
	assert.Equal(t, "vscode", order[1])
	assert.Equal(t, "chrome", order[2])
}

func testEchoArgs(args ...string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("cmd", "/c", "echo", strings.Join(args, " "))
	}
	return exec.Command("/bin/echo", args...)
}

func testTrueCmd() *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("cmd", "/c", "cd")
	}
	return exec.Command("/usr/bin/true")
}

func testFalseCmd() *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("cmd", "/c", "exit", "1")
	}
	return exec.Command("/usr/bin/false")
}

func TestWaitForReady_Success(t *testing.T) {
	saved := execCommandContext
	failCount := 2
	execCommandContext = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		failCount--
		if failCount >= 0 {
			return testFalseCmd()
		}
		return testTrueCmd()
	}
	defer func() { execCommandContext = saved }()

	waitForReady(context.Background(), config.Service{
		Provider: "chrome",
		ReadyCheck: &config.ReadyCheck{
			Cmd: "true", IntervalMs: 10, TimeoutMs: 5000,
		},
	})
}

func TestWaitForReady_Defaults(t *testing.T) {
	saved := execCommandContext
	execCommandContext = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		return testTrueCmd()
	}
	defer func() { execCommandContext = saved }()

	waitForReady(context.Background(), config.Service{
		Provider: "chrome",
		ReadyCheck: &config.ReadyCheck{
			Cmd: "true", IntervalMs: 0, TimeoutMs: 0,
		},
	})
}

func TestWaitDockerHealthy_AllHealthy(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("echo-based output mocking not supported on Windows")
	}
	saved := execCommandContext
	jsonOut := `{"ID":"abc","Health":""}
{"ID":"def","Health":"healthy"}
{"ID":"ghi","Health":"healthy"}`
	execCommandContext = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		return testEchoArgs(jsonOut)
	}
	defer func() { execCommandContext = saved }()

	err := waitDockerHealthy(context.Background())
	require.NoError(t, err)
}

func TestWaitDockerHealthy_StartingThenHealthy(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("echo-based output mocking not supported on Windows")
	}
	callCount := 0
	saved := execCommandContext
	execCommandContext = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		callCount++
		if callCount <= 2 {
			return testEchoArgs(`{"ID":"abc","Health":"starting"}`)
		}
		return testEchoArgs(`{"ID":"abc","Health":"healthy"}`)
	}
	defer func() { execCommandContext = saved }()

	err := waitDockerHealthy(context.Background())
	require.NoError(t, err)
	assert.Greater(t, callCount, 1)
}

func TestWaitDockerHealthy_EmptyOutput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("echo-based output mocking not supported on Windows")
	}
	saved := execCommandContext
	execCommandContext = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		return testEchoArgs("")
	}
	defer func() { execCommandContext = saved }()

	err := waitDockerHealthy(context.Background())
	require.NoError(t, err)
}

func TestServiceLabel(t *testing.T) {
	assert.Equal(t, "Docker Compose", serviceLabel(config.Service{Provider: "docker"}))
	assert.Equal(t, "VS Code", serviceLabel(config.Service{Provider: "vscode"}))
	assert.Equal(t, "DBeaver", serviceLabel(config.Service{Provider: "dbeaver"}))
	assert.Equal(t, "Chrome", serviceLabel(config.Service{Provider: "chrome"}))
	assert.Equal(t, "Command", serviceLabel(config.Service{Provider: "command"}))
	assert.Equal(t, "", serviceLabel(config.Service{Provider: "???"}))
}

func TestServiceSummary(t *testing.T) {
	assert.NotEmpty(t, serviceSummary(config.Service{Provider: "vscode", Folder: "/proj"}))
	assert.NotEmpty(t, serviceSummary(config.Service{Provider: "docker", File: "dc.yml"}))
	assert.NotEmpty(t, serviceSummary(config.Service{Provider: "command", Cmd: "echo"}))
	assert.NotEmpty(t, serviceSummary(config.Service{Provider: "dbeaver", Connection: "conn"}))
	assert.NotEmpty(t, serviceSummary(config.Service{Provider: "chrome", URLs: []string{"url"}}))
}
