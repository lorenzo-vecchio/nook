package provider

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lorenzo-vecchio/nook/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDBeaverProvider_Name(t *testing.T) {
	p := &DBeaverProvider{}
	assert.Equal(t, "dbeaver", p.Name())
}

func TestDBeaverProvider_Detect(t *testing.T) {
	p := &DBeaverProvider{}
	found, err := p.Detect()
	assert.NoError(t, err)
	t.Logf("dbeaver detected: %v", found)
}

func TestDBeaverProvider_Launch(t *testing.T) {
	var capturedName string
	var capturedArgs []string
	saved := execCommandContext
	execCommandContext = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		capturedName = name
		capturedArgs = arg
		return exec.CommandContext(ctx, "/bin/echo", arg...)
	}
	defer func() { execCommandContext = saved }()

	p := &DBeaverProvider{}
	svc := config.Service{
		Provider:   "dbeaver",
		Connection: "jdbc:postgresql://localhost:5432/mydb?user=admin",
	}

	err := p.Launch(context.Background(), svc, "", nil)
	require.NoError(t, err)

	assert.Contains(t, capturedName, "dbeaver")
	assert.Equal(t, []string{"-con", "jdbc:postgresql://localhost:5432/mydb?user=admin"}, capturedArgs)
}

func TestDBeaverProvider_FindPath(t *testing.T) {
	p := &DBeaverProvider{}
	path := p.findDBeaverPath()
	assert.Contains(t, path, "dbeaver")
}

func TestDBeaverCommonPaths(t *testing.T) {
	paths := dbeaverCommonPaths()
	if runtime.GOOS == "darwin" {
		require.Len(t, paths, 1)
		assert.Contains(t, paths[0], "DBeaver.app")
	}
}

func TestDBeaverProvider_DetectInPath(t *testing.T) {
	savedPath := os.Getenv("PATH")
	defer os.Setenv("PATH", savedPath)

	tmpDir := t.TempDir()
	dbeaverPath := filepath.Join(tmpDir, "dbeaver")
	err := os.WriteFile(dbeaverPath, []byte("#!/bin/sh\necho dbeaver"), 0755)
	require.NoError(t, err)

	os.Setenv("PATH", tmpDir)

	p := &DBeaverProvider{}
	found, err := p.Detect()
	assert.NoError(t, err)
	assert.True(t, found)
}

func TestDBeaverProvider_FindPathInLookPath(t *testing.T) {
	savedPath := os.Getenv("PATH")
	defer os.Setenv("PATH", savedPath)

	tmpDir := t.TempDir()
	dbeaverPath := filepath.Join(tmpDir, "dbeaver")
	err := os.WriteFile(dbeaverPath, []byte("#!/bin/sh\necho dbeaver"), 0755)
	require.NoError(t, err)

	os.Setenv("PATH", tmpDir)

	p := &DBeaverProvider{}
	path := p.findDBeaverPath()
	assert.Equal(t, dbeaverPath, path)
}

func TestDBeaverProvider_DetectNotFound(t *testing.T) {
	savedPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", savedPath)

	p := &DBeaverProvider{}
	found, err := p.Detect()
	assert.NoError(t, err)
	t.Logf("dbeaver detected (with empty PATH): %v", found)
}

func TestDBeaverProvider_FindPathNoLookPath(t *testing.T) {
	savedPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", savedPath)

	p := &DBeaverProvider{}
	path := p.findDBeaverPath()
	assert.Contains(t, path, "dbeaver")
}
