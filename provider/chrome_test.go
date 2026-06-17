package provider

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/anomalyco/nook/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChromeProvider_Name(t *testing.T) {
	p := &ChromeProvider{}
	assert.Equal(t, "chrome", p.Name())
}

func TestChromeProvider_Detect(t *testing.T) {
	p := &ChromeProvider{}
	found, err := p.Detect()
	assert.NoError(t, err)
	t.Logf("chrome detected: %v", found)
}

func TestChromeProvider_DetectNotFound(t *testing.T) {
	savedPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", savedPath)

	p := &ChromeProvider{}
	found, err := p.Detect()
	assert.NoError(t, err)
	t.Logf("chrome detected (with empty PATH): %v", found)
}

func TestChromeProvider_LaunchCmdStartFails(t *testing.T) {
	saved := execCommandContext
	execCommandContext = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "nonexistent-binary-xyzzy")
	}
	defer func() { execCommandContext = saved }()

	p := &ChromeProvider{detectedPath: "/usr/bin/google-chrome"}
	svc := config.Service{
		Provider: "chrome",
		URLs:     []string{"https://example.com"},
	}
	err := p.Launch(context.Background(), svc, "", nil)
	assert.Error(t, err)
}

func TestChromeProvider_LaunchAutoDetect(t *testing.T) {
	var captures []struct {
		name string
		args []string
	}
	saved := execCommandContext
	execCommandContext = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		captures = append(captures, struct {
			name string
			args []string
		}{name, arg})
		return exec.CommandContext(ctx, "/bin/echo", arg...)
	}
	defer func() { execCommandContext = saved }()

	savedPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", savedPath)

	p := &ChromeProvider{}

	svc := config.Service{
		Provider: "chrome",
		URLs:     []string{"https://example.com"},
	}

	err := p.Launch(context.Background(), svc, "", nil)
	if err != nil {
		assert.Contains(t, err.Error(), "chrome not found")
		return
	}

	require.Len(t, captures, 1)
	if runtime.GOOS != "windows" {
		assert.Contains(t, captures[0].name, "Google Chrome")
	}
}

func TestChromeProvider_Launch(t *testing.T) {
	var captures []struct {
		name string
		args []string
	}
	saved := execCommandContext
	execCommandContext = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		captures = append(captures, struct {
			name string
			args []string
		}{name, arg})
		return exec.CommandContext(ctx, "/bin/echo", arg...)
	}
	defer func() { execCommandContext = saved }()

	p := &ChromeProvider{detectedPath: "/usr/bin/google-chrome"}

	svc := config.Service{
		Provider: "chrome",
		URLs:     []string{"https://example.com", "https://test.com"},
	}

	err := p.Launch(context.Background(), svc, "", nil)
	require.NoError(t, err)

	require.Len(t, captures, 2)

	if runtime.GOOS == "windows" {
		assert.Equal(t, "cmd", captures[0].name)
		assert.Equal(t, []string{"/c", "start", "chrome", "https://example.com"}, captures[0].args)
		assert.Equal(t, "cmd", captures[1].name)
		assert.Equal(t, []string{"/c", "start", "chrome", "https://test.com"}, captures[1].args)
	} else {
		assert.Equal(t, "/usr/bin/google-chrome", captures[0].name)
		assert.Equal(t, []string{"--new-tab", "https://example.com"}, captures[0].args)
		assert.Equal(t, "/usr/bin/google-chrome", captures[1].name)
		assert.Equal(t, []string{"--new-tab", "https://test.com"}, captures[1].args)
	}
}

func TestChromeProvider_LaunchNoURLs(t *testing.T) {
	saved := execCommandContext
	execCommandContext = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "/bin/echo", arg...)
	}
	defer func() { execCommandContext = saved }()

	p := &ChromeProvider{detectedPath: "/usr/bin/google-chrome"}
	svc := config.Service{Provider: "chrome"}

	err := p.Launch(context.Background(), svc, "", nil)
	assert.NoError(t, err)
}

func TestChromeProvider_DetectAlternativeBinaryName(t *testing.T) {
	savedPath := os.Getenv("PATH")
	defer os.Setenv("PATH", savedPath)

	tmpDir := t.TempDir()
	altBinary := filepath.Join(tmpDir, "chromium-browser")
	err := os.WriteFile(altBinary, []byte("#!/bin/sh\necho chromium"), 0755)
	require.NoError(t, err)

	os.Setenv("PATH", tmpDir)

	p := &ChromeProvider{}
	found, err := p.Detect()
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Contains(t, p.detectedPath, "chromium-browser")
}

func TestChromeBinaryNames(t *testing.T) {
	names := chromeBinaryNames()
	assert.NotEmpty(t, names)

	if runtime.GOOS == "windows" {
		assert.Equal(t, "chrome.exe", names[0])
	} else {
		assert.Contains(t, names, "google-chrome")
	}
}

func TestChromePaths(t *testing.T) {
	paths := chromePaths()
	assert.NotEmpty(t, paths)

	if runtime.GOOS == "darwin" {
		assert.Contains(t, paths[0], "Google Chrome.app")
	}
	if runtime.GOOS == "linux" {
		assert.Contains(t, paths[0], "google-chrome")
	}
	if runtime.GOOS == "windows" {
		assert.Contains(t, paths[0], "chrome.exe")
	}
}

func TestChromeCommand(t *testing.T) {
	ctx := context.Background()

	if runtime.GOOS == "windows" {
		cmd := chromeCommand(ctx, "", "https://example.com")
		assert.Equal(t, "cmd", cmd.Path)
		assert.Equal(t, []string{"cmd", "/c", "start", "chrome", "https://example.com"}, cmd.Args)
	} else {
		cmd := chromeCommand(ctx, "/usr/bin/google-chrome", "https://example.com")
		assert.Equal(t, "/usr/bin/google-chrome", cmd.Path)
		assert.Equal(t, []string{"/usr/bin/google-chrome", "--new-tab", "https://example.com"}, cmd.Args)
	}
}
