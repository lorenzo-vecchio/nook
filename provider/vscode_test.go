package provider

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/anomalyco/nook/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVSCodeProvider_Name(t *testing.T) {
	p := &VSCodeProvider{}
	assert.Equal(t, "vscode", p.Name())
}

func TestVSCodeProvider_Detect(t *testing.T) {
	p := &VSCodeProvider{}
	found, err := p.Detect()
	assert.NoError(t, err)
	t.Logf("vscode detected: %v", found)
}

func TestVSCodeProvider_LaunchNoTerminals(t *testing.T) {
	var capturedName string
	var capturedArgs []string
	saved := execCommandContext
	execCommandContext = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		capturedName = name
		capturedArgs = arg
		return exec.CommandContext(ctx, "/bin/echo", arg...)
	}
	defer func() { execCommandContext = saved }()

	p := &VSCodeProvider{}
	tmpDir := t.TempDir()

	svc := config.Service{
		Provider: "vscode",
		Folder:   "myproject",
	}

	err := p.Launch(context.Background(), svc, tmpDir, nil)
	require.NoError(t, err)

	assert.Equal(t, "code", capturedName)
	assert.Len(t, capturedArgs, 1)
	expectedFolder := filepath.Join(tmpDir, "myproject")
	assert.Equal(t, expectedFolder, capturedArgs[0])
}

func TestVSCodeProvider_LaunchWithTerminals(t *testing.T) {
	var capturedName string
	var capturedArgs []string
	saved := execCommandContext
	execCommandContext = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		capturedName = name
		capturedArgs = arg
		return exec.CommandContext(ctx, "/bin/echo", arg...)
	}
	defer func() { execCommandContext = saved }()

	p := &VSCodeProvider{}
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "myproject")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	svc := config.Service{
		Provider: "vscode",
		Folder:   "myproject",
		Terminals: []config.Terminal{
			{Name: "dev", Directory: ".", Command: "npm run dev"},
		},
	}

	err = p.Launch(context.Background(), svc, tmpDir, nil)
	require.NoError(t, err)

	assert.Equal(t, "code", capturedName)
	require.Len(t, capturedArgs, 1)
	assert.Contains(t, capturedArgs[0], ".code-workspace")
}

func TestVSCodeProvider_GenerateWorkspaceFile(t *testing.T) {
	p := &VSCodeProvider{}
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "myproject")

	svc := config.Service{
		Provider: "vscode",
		Folder:   "myproject",
		Terminals: []config.Terminal{
			{Name: "dev", Directory: ".", Command: "npm run dev"},
			{Name: "build", Directory: ".", Command: "npm run build"},
		},
	}

	wsPath, err := p.generateWorkspaceFile(
		projectDir,
		svc.Terminals,
		tmpDir,
	)
	require.NoError(t, err)

	_, err = os.Stat(wsPath)
	assert.NoError(t, err)

	data, err := os.ReadFile(wsPath)
	require.NoError(t, err)

	var ws codeWorkspace
	err = json.Unmarshal(data, &ws)
	require.NoError(t, err)

	require.Len(t, ws.Folders, 1)
	assert.Equal(t, projectDir, ws.Folders[0].Path)

	require.Len(t, ws.Tasks.Tasks, 2)
	assert.Equal(t, "dev", ws.Tasks.Tasks[0].Label)
	assert.Equal(t, "build", ws.Tasks.Tasks[1].Label)
	assert.Equal(t, "folderOpen", ws.Tasks.Tasks[0].RunOn)
	assert.Equal(t, "shell", ws.Tasks.Tasks[0].Type)
	assert.Empty(t, ws.Tasks.Tasks[0].ProblemMatcher)
}

func TestVSCodeCommonPaths(t *testing.T) {
	paths := vscodeCommonPaths()
	assert.NotEmpty(t, paths)
}

func TestVSCodeProvider_DetectNotFound(t *testing.T) {
	savedPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", savedPath)

	p := &VSCodeProvider{}
	found, err := p.Detect()
	assert.NoError(t, err)
	t.Logf("vscode detected (with empty PATH): %v", found)
}

func TestVSCodeProvider_GenerateWorkspaceFileEnsureDirFails(t *testing.T) {
	p := &VSCodeProvider{}
	tmpDir := t.TempDir()

	filePath := filepath.Join(tmpDir, ".workspace")
	err := os.WriteFile(filePath, []byte(""), 0644)
	require.NoError(t, err)

	projectDir := filepath.Join(tmpDir, "myproject")
	err = os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	terminals := []config.Terminal{
		{Name: "dev", Directory: ".", Command: "npm run dev"},
	}

	_, err = p.generateWorkspaceFile(projectDir, terminals, tmpDir)
	assert.Error(t, err)
}

func TestVSCodeProvider_LaunchWithTerminalsGenerateWSFails(t *testing.T) {
	saved := execCommandContext
	execCommandContext = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "/bin/echo", arg...)
	}
	defer func() { execCommandContext = saved }()

	p := &VSCodeProvider{}
	tmpDir := t.TempDir()

	filePath := filepath.Join(tmpDir, ".workspace")
	err := os.WriteFile(filePath, []byte(""), 0644)
	require.NoError(t, err)

	svc := config.Service{
		Provider: "vscode",
		Folder:   "myproject",
		Terminals: []config.Terminal{
			{Name: "dev", Directory: ".", Command: "npm run dev"},
		},
	}

	projectDir := filepath.Join(tmpDir, "myproject")
	err = os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	err = p.Launch(context.Background(), svc, tmpDir, nil)
	assert.Error(t, err)
}

func TestVSCodeProvider_GenerateWorkspaceFileWriteFails(t *testing.T) {
	p := &VSCodeProvider{}
	tmpDir := t.TempDir()

	projectDir := filepath.Join(tmpDir, "myproject")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	wsDir := filepath.Join(tmpDir, ".workspace")
	err = os.MkdirAll(wsDir, 0555)
	require.NoError(t, err)

	terminals := []config.Terminal{
		{Name: "dev", Directory: ".", Command: "npm run dev"},
	}

	_, err = p.generateWorkspaceFile(projectDir, terminals, tmpDir)
	assert.Error(t, err)
}

func TestVSCodeProvider_DetectCodeInsiders(t *testing.T) {
	savedPath := os.Getenv("PATH")
	defer os.Setenv("PATH", savedPath)

	tmpDir := t.TempDir()
	codeInsidersPath := filepath.Join(tmpDir, "code-insiders")
	err := os.WriteFile(codeInsidersPath, []byte("#!/bin/sh\necho code-insiders"), 0755)
	require.NoError(t, err)

	os.Setenv("PATH", tmpDir)

	p := &VSCodeProvider{}
	found, err := p.Detect()
	assert.NoError(t, err)
	assert.True(t, found)
}

func TestVSCodeProvider_GenerateWorkspaceFileTerminalNoCommand(t *testing.T) {
	p := &VSCodeProvider{}
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "myproject")

	terminals := []config.Terminal{
		{Name: "shell", Directory: "."},
	}

	wsPath, err := p.generateWorkspaceFile(projectDir, terminals, tmpDir)
	require.NoError(t, err)

	data, err := os.ReadFile(wsPath)
	require.NoError(t, err)

	var ws codeWorkspace
	err = json.Unmarshal(data, &ws)
	require.NoError(t, err)

	require.Len(t, ws.Tasks.Tasks, 1)
	assert.Equal(t, "cd "+projectDir, ws.Tasks.Tasks[0].Command)
}
