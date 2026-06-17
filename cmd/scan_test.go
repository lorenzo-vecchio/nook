package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/anomalyco/nook/config"
)

func TestNewScanCmd(t *testing.T) {
	cmd := NewScanCmd()
	assert.Equal(t, "scan", cmd.Use)
}

func TestScanCmd_Run(t *testing.T) {
	cmd := NewScanCmd()
	cmd.SetArgs([]string{})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestScanCmd_NonExistentPath(t *testing.T) {
	origCfg, err := config.LoadGlobalConfig()
	assert.NoError(t, err)
	savedPaths := origCfg.ScanPaths
	defer func() {
		origCfg.ScanPaths = savedPaths
		config.SaveGlobalConfig(origCfg)
	}()

	origCfg.ScanPaths = []string{"/tmp/nonexistent_scan_path_12345"}
	config.SaveGlobalConfig(origCfg)

	cmd := NewScanCmd()
	cmd.SetArgs([]string{})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err = cmd.Execute()
	assert.NoError(t, err)

	cfg, err := config.LoadGlobalConfig()
	assert.NoError(t, err)
	assert.NotContains(t, cfg.ScanPaths, "/tmp/nonexistent_scan_path_12345")
}

func TestScanCmd_WithWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	wsDir := filepath.Join(tmpDir, "myworkspace")
	err := os.MkdirAll(wsDir, 0755)
	assert.NoError(t, err)

	wsContent := []byte("name: myworkspace\ndescription: test workspace\nenvironments:\n  dev:\n    services:\n      - provider: vscode\n")
	err = os.WriteFile(filepath.Join(wsDir, "workspace.yaml"), wsContent, 0644)
	assert.NoError(t, err)

	origCfg, err := config.LoadGlobalConfig()
	assert.NoError(t, err)
	savedPaths := origCfg.ScanPaths
	defer func() {
		origCfg.ScanPaths = savedPaths
		config.SaveGlobalConfig(origCfg)
	}()

	origCfg.ScanPaths = []string{tmpDir}
	config.SaveGlobalConfig(origCfg)

	cmd := NewScanCmd()
	cmd.SetArgs([]string{})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err = cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "myworkspace")
}
