package provider

import (
	"context"
	"os/exec"
	"testing"

	"github.com/anomalyco/nook/config"
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
