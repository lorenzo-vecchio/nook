package detector

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lorenzo-vecchio/nook/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeWorkspaceYAML(t *testing.T, dir, name, desc string) {
	t.Helper()
	content := "name: " + name + "\n" +
		"description: " + desc + "\n" +
		"environments:\n" +
		"  dev:\n" +
		"    services:\n" +
		"      - provider: command\n" +
		"        cmd: \"echo hello\"\n"
	path := filepath.Join(dir, "workspace.yaml")
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
}

func TestScanCurrentDir_NoWorkspaceYAML(t *testing.T) {
	dir := t.TempDir()
	origWd, _ := os.Getwd()
	err := os.Chdir(dir)
	require.NoError(t, err)
	defer func() { _ = os.Chdir(origWd) }()

	result, err := ScanCurrentDir()
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.InCWD)
	assert.Empty(t, result.InSubdirs)
	assert.False(t, result.HasNew)
}

func TestScanCurrentDir_WorkspaceInCWD(t *testing.T) {
	dir := t.TempDir()
	realDir, err := filepath.EvalSymlinks(dir)
	require.NoError(t, err)
	writeWorkspaceYAML(t, realDir, "test-ws", "A test workspace")
	origWd, _ := os.Getwd()
	err = os.Chdir(realDir)
	require.NoError(t, err)
	defer func() { _ = os.Chdir(origWd) }()

	result, err := ScanCurrentDir()
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.InCWD, 1)
	assert.Equal(t, "test-ws", result.InCWD[0].Name)
	assert.Equal(t, "A test workspace", result.InCWD[0].Description)
	assert.Equal(t, realDir, result.InCWD[0].Path)
	assert.Empty(t, result.InSubdirs)
	assert.True(t, result.HasNew)
}

func TestScanCurrentDir_WorkspaceInSubdir(t *testing.T) {
	dir := t.TempDir()
	realDir, err := filepath.EvalSymlinks(dir)
	require.NoError(t, err)
	subdir := filepath.Join(realDir, "my-project")
	err = os.MkdirAll(subdir, 0755)
	require.NoError(t, err)
	writeWorkspaceYAML(t, subdir, "sub-ws", "Subdir workspace")

	origWd, _ := os.Getwd()
	err = os.Chdir(realDir)
	require.NoError(t, err)
	defer func() { _ = os.Chdir(origWd) }()

	result, err := ScanCurrentDir()
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.InCWD)
	assert.Len(t, result.InSubdirs, 1)
	assert.Equal(t, "sub-ws", result.InSubdirs[0].Name)
	assert.Equal(t, "Subdir workspace", result.InSubdirs[0].Description)
	assert.Equal(t, subdir, result.InSubdirs[0].Path)
	assert.True(t, result.HasNew)
}

func TestScanCurrentDir_WorkspaceInBoth(t *testing.T) {
	dir := t.TempDir()
	realDir, err := filepath.EvalSymlinks(dir)
	require.NoError(t, err)
	subdir := filepath.Join(realDir, "sub-project")
	err = os.MkdirAll(subdir, 0755)
	require.NoError(t, err)
	writeWorkspaceYAML(t, realDir, "cwd-ws", "CWD workspace")
	writeWorkspaceYAML(t, subdir, "sub-ws", "Subdir workspace")

	origWd, _ := os.Getwd()
	err = os.Chdir(realDir)
	require.NoError(t, err)
	defer func() { _ = os.Chdir(origWd) }()

	result, err := ScanCurrentDir()
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.InCWD, 1)
	assert.Len(t, result.InSubdirs, 1)
	assert.True(t, result.HasNew)
}

func TestScanPath_FindsWorkspaceYAML(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "a-project")
	err := os.MkdirAll(subdir, 0755)
	require.NoError(t, err)
	writeWorkspaceYAML(t, dir, "root-ws", "Root")
	writeWorkspaceYAML(t, subdir, "a-ws", "A project")

	workspaces, err := ScanPath(dir)
	require.NoError(t, err)
	require.NotNil(t, workspaces)
	assert.Len(t, workspaces, 2)

	root, ok := workspaces["root-ws"]
	assert.True(t, ok)
	assert.Equal(t, "Root", root.Description)
	assert.Equal(t, dir, root.Path)

	ws, ok := workspaces["a-ws"]
	assert.True(t, ok)
	assert.Equal(t, "A project", ws.Description)
	assert.Equal(t, subdir, ws.Path)
}

func TestScanPath_NoFiles(t *testing.T) {
	dir := t.TempDir()
	workspaces, err := ScanPath(dir)
	require.NoError(t, err)
	assert.Empty(t, workspaces)
}

func TestScanPath_NonExistentDir(t *testing.T) {
	_, err := ScanPath("/nonexistent/path/12345")
	assert.Error(t, err)
}

func TestIsTrusted_ReturnsTrue(t *testing.T) {
	assert.True(t, IsTrusted("/home/user/projects", []string{"/home/user/projects"}))
	assert.True(t, IsTrusted("/home/user/projects/sub", []string{"/home/user/projects"}))
}

func TestIsTrusted_ReturnsFalse(t *testing.T) {
	assert.False(t, IsTrusted("/other/path", []string{"/home/user/projects"}))
	assert.False(t, IsTrusted("/home/user", []string{"/home/user/projects"}))
}

func TestIsTrusted_EmptyScanPaths(t *testing.T) {
	assert.False(t, IsTrusted("/any/path", []string{}))
}

func TestIsTrusted_UsesAbsPath(t *testing.T) {
	result := IsTrusted("relative/path", []string{})
	assert.False(t, result)
}

func TestTrustPath_AddsPathToConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg, err := config.LoadGlobalConfig()
	require.NoError(t, err)
	assert.Len(t, cfg.ScanPaths, 1)

	err = TrustPath("/new/trusted/path")
	require.NoError(t, err)

	cfg, err = config.LoadGlobalConfig()
	require.NoError(t, err)
	assert.Contains(t, cfg.ScanPaths, "/new/trusted/path")
}

func TestScanCurrentDir_DirNamedWorkspaceYAML(t *testing.T) {
	dir := t.TempDir()
	realDir, err := filepath.EvalSymlinks(dir)
	require.NoError(t, err)
	badDir := filepath.Join(realDir, "workspace.yaml")
	err = os.MkdirAll(badDir, 0755)
	require.NoError(t, err)

	origWd, _ := os.Getwd()
	err = os.Chdir(realDir)
	require.NoError(t, err)
	defer func() { _ = os.Chdir(origWd) }()

	result, err := ScanCurrentDir()
	require.NoError(t, err)
	assert.Empty(t, result.InCWD)
	assert.False(t, result.HasNew)
}

func TestTrustPath_DuplicatePath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	err := TrustPath("/some/path")
	require.NoError(t, err)

	err = TrustPath("/some/path")
	require.NoError(t, err)

	cfg, err := config.LoadGlobalConfig()
	require.NoError(t, err)
	n := 0
	for _, p := range cfg.ScanPaths {
		if p == "/some/path" {
			n++
		}
	}
	assert.Equal(t, 1, n)
}

func TestIsTrusted_SubdirOfScanPath(t *testing.T) {
	assert.True(t, IsTrusted("/home/user/projects/my-app", []string{"/home/user/projects"}))
}

func TestIsTrusted_ParentOfScanPathNotTrusted(t *testing.T) {
	assert.False(t, IsTrusted("/home/user", []string{"/home/user/projects"}))
}

func TestScanCurrentDir_InvalidWorkspaceYAML(t *testing.T) {
	dir := t.TempDir()
	realDir, err := filepath.EvalSymlinks(dir)
	require.NoError(t, err)
	badPath := filepath.Join(realDir, "workspace.yaml")
	err = os.WriteFile(badPath, []byte("invalid: [yaml"), 0644)
	require.NoError(t, err)

	origWd, _ := os.Getwd()
	err = os.Chdir(realDir)
	require.NoError(t, err)
	defer func() { _ = os.Chdir(origWd) }()

	_, err = ScanCurrentDir()
	assert.Error(t, err)
}
