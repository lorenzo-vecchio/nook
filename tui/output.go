package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
)

type WorkspaceListItem struct {
	Name        string
	Description string
	Envs        []string
	Path        string
}

func PrintProgress(w io.Writer, providerName string, message string) {
	provider := color.New(color.FgCyan, color.Bold).Sprintf("  [%s]", providerName)
	fmt.Fprintf(w, "%s    %s\n", provider, message)
}

func PrintSuccess(w io.Writer, message string) {
	fmt.Fprintf(w, "%s %s\n", color.GreenString("✔"), message)
}

func PrintError(w io.Writer, message string) {
	fmt.Fprintf(w, "%s %s\n", color.RedString("✖"), message)
}

func PrintHeader(w io.Writer, header string) {
	s := color.New(color.FgBlue, color.Bold).Sprintf("=== %s ===", header)
	fmt.Fprintln(w, s)
}

func PrintDetectionTable(w io.Writer, results map[string]bool) {
	header := color.New(color.FgBlue, color.Bold).Sprintf("  %-20s %s", "Provider", "Status")
	fmt.Fprintln(w, header)
	fmt.Fprintln(w, "  "+strings.Repeat("─", 30))
	for provider, detected := range results {
		var status string
		if detected {
			status = color.GreenString("✔ detected")
		} else {
			status = color.RedString("✗ not found")
		}
		fmt.Fprintf(w, "  %-20s %s\n", provider, status)
	}
}

func PrintWorkspaceList(w io.Writer, workspaces []WorkspaceListItem) {
	if len(workspaces) == 0 {
		fmt.Fprintln(w, "No workspaces found")
		return
	}
	header := color.New(color.FgBlue, color.Bold).Sprintf("  %-20s %-30s %s", "Name", "Description", "Path")
	fmt.Fprintln(w, header)
	fmt.Fprintln(w, "  "+strings.Repeat("─", 80))
	for _, ws := range workspaces {
		name := color.CyanString(ws.Name)
		desc := ws.Description
		if desc == "" {
			desc = "-"
		}
		path := color.New(color.Faint).Sprint(ws.Path)
		fmt.Fprintf(w, "  %-20s %-30s %s\n", name, desc, path)
	}
}
