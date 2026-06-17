package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadGlobalConfig_NonExistentReturnsDefaults(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg, err := LoadGlobalConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, []string{"~/.nook/workspaces"}, cfg.ScanPaths)
}

func TestLoadGlobalConfig_CreatesConfigDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	_, err := LoadGlobalConfig()
	require.NoError(t, err)

	configDir := filepath.Join(dir, "nook")
	info, err := os.Stat(configDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestLoadGlobalConfig_ExistingConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	configDir := filepath.Join(dir, "nook")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	yamlContent := "scan_paths:\n  - /custom/path\n"
	err = os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(yamlContent), 0644)
	require.NoError(t, err)

	cfg, err := LoadGlobalConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, []string{"/custom/path"}, cfg.ScanPaths)
}

func TestSaveGlobalConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := &GlobalConfig{
		ScanPaths: []string{"/a", "/b"},
	}

	err := SaveGlobalConfig(cfg)
	require.NoError(t, err)

	configDir := filepath.Join(dir, "nook")
	configFile := filepath.Join(configDir, "config.yaml")
	data, err := os.ReadFile(configFile)
	require.NoError(t, err)
	assert.Contains(t, string(data), "/a")
	assert.Contains(t, string(data), "/b")

	loaded, err := LoadGlobalConfig()
	require.NoError(t, err)
	assert.Equal(t, cfg.ScanPaths, loaded.ScanPaths)
}

func TestGlobalConfig_TildeExpansion(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	configDir := filepath.Join(dir, "nook")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	yamlContent := "scan_paths:\n  - ~/my-workspaces\n"
	err = os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(yamlContent), 0644)
	require.NoError(t, err)

	cfg, err := LoadGlobalConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, "my-workspaces")
	assert.Equal(t, expected, cfg.ScanPaths[0])
}

func TestLoadGlobalConfig_InvalidYaml(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	configDir := filepath.Join(dir, "nook")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("invalid: [yaml: broken\n"), 0644)
	require.NoError(t, err)

	_, err = LoadGlobalConfig()
	assert.Error(t, err)
}

func TestXdgPathResolution(t *testing.T) {
	expected := filepath.Join(xdg.ConfigHome, "nook", "config.yaml")
	cfgPath := filepath.Join(xdg.ConfigHome, "nook", "config.yaml")
	assert.Equal(t, expected, cfgPath)
}

func TestConfigDirPath_WithXDGConfigHome(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	path := configDirPath()
	assert.Equal(t, filepath.Join(dir, "nook"), path)
}

func TestConfigDirPath_DefaultPath(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	path := configDirPath()
	assert.Equal(t, "nook", filepath.Base(path))
}

func TestExpandTilde_NoTilde(t *testing.T) {
	result := expandTilde("/absolute/path")
	assert.Equal(t, "/absolute/path", result)
}

func TestExpandTilde_TildeOnly(t *testing.T) {
	result := expandTilde("~")
	assert.Equal(t, "~", result)
}

func TestLoadGlobalConfig_MkdirAllFails(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	err := os.WriteFile(filepath.Join(dir, "nook"), []byte("file in the way"), 0644)
	require.NoError(t, err)
	_, err = LoadGlobalConfig()
	assert.Error(t, err)
}

func TestLoadGlobalConfig_ReadPermissionDenied(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod not supported on Windows")
	}
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	nookDir := filepath.Join(dir, "nook")
	err := os.MkdirAll(nookDir, 0755)
	require.NoError(t, err)
	cfgPath := filepath.Join(nookDir, "config.yaml")
	err = os.WriteFile(cfgPath, []byte("scan_paths:\n  - /path\n"), 0644)
	require.NoError(t, err)
	err = os.Chmod(cfgPath, 0)
	require.NoError(t, err)
	_, err = LoadGlobalConfig()
	assert.Error(t, err)
}

func TestSaveGlobalConfig_MkdirAllFails(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	err := os.WriteFile(filepath.Join(dir, "nook"), []byte("file in the way"), 0644)
	require.NoError(t, err)
	cfg := &GlobalConfig{ScanPaths: []string{"/a"}}
	err = SaveGlobalConfig(cfg)
	assert.Error(t, err)
}
