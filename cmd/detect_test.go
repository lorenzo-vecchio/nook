package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDetectCmd(t *testing.T) {
	cmd := NewDetectCmd()
	assert.Equal(t, "detect", cmd.Use)
}

func TestDetectCmd_Run(t *testing.T) {
	cmd := NewDetectCmd()
	cmd.SetArgs([]string{})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Provider")
}
