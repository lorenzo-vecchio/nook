package tui

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrintProgress(t *testing.T) {
	var buf bytes.Buffer
	PrintProgress(&buf, "vscode", "Opening workspace")
	assert.Contains(t, buf.String(), "vscode")
	assert.Contains(t, buf.String(), "Opening workspace")
}

func TestPrintSuccess(t *testing.T) {
	var buf bytes.Buffer
	PrintSuccess(&buf, "Done!")
	assert.Contains(t, buf.String(), "✔")
	assert.Contains(t, buf.String(), "Done!")
}

func TestPrintError(t *testing.T) {
	var buf bytes.Buffer
	PrintError(&buf, "Something failed")
	assert.Contains(t, buf.String(), "✖")
	assert.Contains(t, buf.String(), "Something failed")
}

func TestPrintHeader(t *testing.T) {
	var buf bytes.Buffer
	PrintHeader(&buf, "My Header")
	assert.Contains(t, buf.String(), "My Header")
}

func TestPrintDetectionTable(t *testing.T) {
	var buf bytes.Buffer
	results := map[string]bool{
		"vscode":  true,
		"dbeaver": false,
	}
	PrintDetectionTable(&buf, results)
	output := buf.String()
	assert.Contains(t, output, "vscode")
	assert.Contains(t, output, "dbeaver")
}

func TestPrintWorkspaceList(t *testing.T) {
	var buf bytes.Buffer
	workspaces := []WorkspaceListItem{
		{Name: "myapp", Description: "My app", Envs: []string{"dev", "prod"}, Path: "/path/to/myapp"},
	}
	PrintWorkspaceList(&buf, workspaces)
	output := buf.String()
	assert.Contains(t, output, "myapp")
	assert.Contains(t, output, "My app")
	assert.Contains(t, output, "/path/to/myapp")
}

func TestPrintWorkspaceListEmpty(t *testing.T) {
	var buf bytes.Buffer
	PrintWorkspaceList(&buf, []WorkspaceListItem{})
	assert.Contains(t, buf.String(), "No workspaces found")
}
