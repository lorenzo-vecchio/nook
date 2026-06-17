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

func TestRootHasSubcommands(t *testing.T) {
	assert.True(t, len(rootCmd.Commands()) > 0)
}
