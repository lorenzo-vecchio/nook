package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveEnvVars_FromOS(t *testing.T) {
	t.Setenv("NOOK_TEST_HOME", "/home/user")
	result := ResolveEnvVars("${NOOK_TEST_HOME}/project", nil)
	assert.Equal(t, "/home/user/project", result)
}

func TestResolveEnvVars_FromExtraMap(t *testing.T) {
	extra := map[string]string{"CUSTOM_VAR": "custom_value"}
	result := ResolveEnvVars("prefix-${CUSTOM_VAR}-suffix", extra)
	assert.Equal(t, "prefix-custom_value-suffix", result)
}

func TestResolveEnvVars_ExtraTakesPrecedence(t *testing.T) {
	t.Setenv("SHARED_VAR", "from_os")
	extra := map[string]string{"SHARED_VAR": "from_extra"}
	result := ResolveEnvVars("${SHARED_VAR}", extra)
	assert.Equal(t, "from_extra", result)
}

func TestResolveEnvVars_MissingVarLeavesPlaceholder(t *testing.T) {
	result := ResolveEnvVars("${MISSING_VAR}", nil)
	assert.Equal(t, "${MISSING_VAR}", result)
}

func TestResolveEnvVars_MultipleVars(t *testing.T) {
	t.Setenv("A", "hello")
	t.Setenv("B", "world")
	result := ResolveEnvVars("${A} ${B}", nil)
	assert.Equal(t, "hello world", result)
}

func TestResolveEnvVars_NoPlaceholders(t *testing.T) {
	result := ResolveEnvVars("plain string", nil)
	assert.Equal(t, "plain string", result)
}

func TestResolveEnvVars_EmptyString(t *testing.T) {
	result := ResolveEnvVars("", nil)
	assert.Equal(t, "", result)
}

func TestResolveEnvVars_DollarWithoutBraces(t *testing.T) {
	result := ResolveEnvVars("$NOT_A_VAR", nil)
	assert.Equal(t, "$NOT_A_VAR", result)
}

func TestResolveEnvVars_PartialMatch(t *testing.T) {
	result := ResolveEnvVars("${A}", nil)
	assert.Equal(t, "${A}", result)
}

func TestLoadEnvFile(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	content := `KEY=value
EMPTY=
QUOTED="quoted value"
SPACES=hello world`
	err := os.WriteFile(envPath, []byte(content), 0644)
	require.NoError(t, err)

	env, err := LoadEnvFile(envPath)
	require.NoError(t, err)
	assert.Equal(t, "value", env["KEY"])
	assert.Equal(t, "", env["EMPTY"])
	assert.Equal(t, "quoted value", env["QUOTED"])
	assert.Equal(t, "hello world", env["SPACES"])
}

func TestLoadEnvFile_FileNotFound(t *testing.T) {
	_, err := LoadEnvFile("/nonexistent/.env")
	assert.Error(t, err)
}

func TestResolveAllEnvVars(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	err := os.WriteFile(envPath, []byte("FROM_FILE=file_value"), 0644)
	require.NoError(t, err)

	t.Setenv("FROM_OS", "os_value")
	extra := map[string]string{"FROM_EXTRA": "extra_value"}

	input := "${FROM_OS}-${FROM_EXTRA}-${FROM_FILE}-${MISSING}"
	result := ResolveAllEnvVars(input, envPath, extra)
	assert.Equal(t, "os_value-extra_value-file_value-${MISSING}", result)
}

func TestResolveAllEnvVars_NoEnvFile(t *testing.T) {
	t.Setenv("VAR", "value")
	result := ResolveAllEnvVars("${VAR}", "", nil)
	assert.Equal(t, "value", result)
}

func TestResolveAllEnvVars_EnvFileOverridesOS(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	err := os.WriteFile(envPath, []byte("VAR=file_value"), 0644)
	require.NoError(t, err)

	t.Setenv("VAR", "os_value")
	result := ResolveAllEnvVars("${VAR}", envPath, nil)
	assert.Equal(t, "file_value", result)
}

func TestLoadEnvFile_WithCommentsAndBlankLines(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	content := `# This is a comment

KEY=value
# another comment
NEXT=val2`
	err := os.WriteFile(envPath, []byte(content), 0644)
	require.NoError(t, err)

	env, err := LoadEnvFile(envPath)
	require.NoError(t, err)
	assert.Equal(t, "value", env["KEY"])
	assert.Equal(t, "val2", env["NEXT"])
	assert.Len(t, env, 2)
}
