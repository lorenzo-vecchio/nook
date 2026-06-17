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

func TestCommandProvider_Name(t *testing.T) {
	p := &CommandProvider{}
	assert.Equal(t, "command", p.Name())
}

func TestCommandProvider_Detect(t *testing.T) {
	p := &CommandProvider{}
	ok, err := p.Detect()
	assert.True(t, ok)
	assert.NoError(t, err)
}

func TestCommandProvider_Launch(t *testing.T) {
	var capturedName string
	var capturedArgs []string
	var capturedCmd *exec.Cmd
	saved := execCommandContext
	execCommandContext = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		capturedName = name
		capturedArgs = arg
		capturedCmd = testCmd(ctx, arg...)
		return capturedCmd
	}
	defer func() { execCommandContext = saved }()

	p := &CommandProvider{}
	tmpDir := t.TempDir()
	subdir := filepath.Join(tmpDir, "subdir")
	err := os.MkdirAll(subdir, 0755)
	require.NoError(t, err)

	svc := config.Service{
		Provider: "command",
		Cmd:      "echo hello",
		Cwd:      "subdir",
	}

	err = p.Launch(context.Background(), svc, tmpDir, nil)
	require.NoError(t, err)

	assert.Equal(t, subdir, capturedCmd.Dir)

	if runtime.GOOS == "windows" {
		assert.Equal(t, "cmd", capturedName)
		assert.Equal(t, []string{"/c", "echo hello"}, capturedArgs)
	} else {
		assert.Equal(t, "sh", capturedName)
		assert.Equal(t, []string{"-c", "echo hello"}, capturedArgs)
	}
}

func TestCommandProvider_LaunchNoCwd(t *testing.T) {
	var capturedCmd *exec.Cmd
	saved := execCommandContext
	execCommandContext = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		capturedCmd = testCmd(ctx, arg...)
		return capturedCmd
	}
	defer func() { execCommandContext = saved }()

	p := &CommandProvider{}
	svc := config.Service{
		Provider: "command",
		Cmd:      "echo hello",
	}

	err := p.Launch(context.Background(), svc, "", nil)
	require.NoError(t, err)
	assert.Equal(t, "", capturedCmd.Dir)
}
