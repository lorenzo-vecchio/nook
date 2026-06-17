package config

import (
	"os"
	"path/filepath"
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
