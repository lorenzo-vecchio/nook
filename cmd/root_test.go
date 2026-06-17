package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"--help"})
	err := rootCmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Nook")
}

func TestRootCmdUses(t *testing.T) {
	assert.Equal(t, "nook", rootCmd.Use)
}

func TestRootPersistentPreRunE(t *testing.T) {
	cfg = nil
	err := rootCmd.PersistentPreRunE(rootCmd, []string{})
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
}
