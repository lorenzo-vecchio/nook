package provider

import (
	"context"
	"os/exec"
	"testing"

	"github.com/anomalyco/nook/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDockerProvider_Name(t *testing.T) {
	p := &DockerProvider{}
	assert.Equal(t, "docker", p.Name())
}

func TestDockerProvider_Detect(t *testing.T) {
	p := &DockerProvider{}
	found, err := p.Detect()
	assert.NoError(t, err)
	t.Logf("docker detected: %v", found)
}

func TestDockerProvider_Launch(t *testing.T) {
	var capturedName string
	var capturedArgs []string
	saved := execCommandContext
	execCommandContext = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		capturedName = name
		capturedArgs = arg
		return exec.CommandContext(ctx, "/bin/echo", arg...)
	}
	defer func() { execCommandContext = saved }()

	p := &DockerProvider{}
	tmpDir := t.TempDir()

	svc := config.Service{
		Provider: "docker",
		File:     "docker-compose.yml",
	}

	err := p.Launch(context.Background(), svc, tmpDir, nil)
	require.NoError(t, err)

	assert.Equal(t, "docker", capturedName)
	assert.Equal(t, []string{"compose", "-f", tmpDir + "/docker-compose.yml", "up", "-d"}, capturedArgs)
}

func TestDockerProvider_LaunchWithProfile(t *testing.T) {
	var capturedName string
	var capturedArgs []string
	saved := execCommandContext
	execCommandContext = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		capturedName = name
		capturedArgs = arg
		return exec.CommandContext(ctx, "/bin/echo", arg...)
	}
	defer func() { execCommandContext = saved }()

	p := &DockerProvider{}
	tmpDir := t.TempDir()

	svc := config.Service{
		Provider: "docker",
		File:     "compose.prod.yml",
		Profile:  "production",
	}

	err := p.Launch(context.Background(), svc, tmpDir, nil)
	require.NoError(t, err)

	assert.Equal(t, "docker", capturedName)
	assert.Equal(t, []string{"compose", "-f", tmpDir + "/compose.prod.yml", "--profile", "production", "up", "-d"}, capturedArgs)
}

func TestDockerProvider_LaunchAbsolutePath(t *testing.T) {
	var capturedArgs []string
	saved := execCommandContext
	execCommandContext = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		capturedArgs = arg
		return exec.CommandContext(ctx, "/bin/echo", arg...)
	}
	defer func() { execCommandContext = saved }()

	p := &DockerProvider{}

	svc := config.Service{
		Provider: "docker",
		File:     "/etc/nook/docker-compose.yml",
	}

	err := p.Launch(context.Background(), svc, "/tmp/base", nil)
	require.NoError(t, err)

	assert.Equal(t, []string{"compose", "-f", "/etc/nook/docker-compose.yml", "up", "-d"}, capturedArgs)
}
