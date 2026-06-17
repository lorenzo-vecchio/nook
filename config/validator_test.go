package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate_ValidGlobalConfig(t *testing.T) {
	cfg := &GlobalConfig{
		ScanPaths: []string{"/some/path"},
	}
	err := Validate(cfg)
	assert.NoError(t, err)
}

func TestValidate_ValidWorkspaceConfig(t *testing.T) {
	ws := &WorkspaceConfig{
		Name: "my-project",
		Environments: map[string]Environment{
			"dev": {
				Services: []Service{
					{Provider: "vscode", Folder: "/path"},
				},
			},
		},
	}
	err := Validate(ws)
	assert.NoError(t, err)
}

func TestValidate_MissingName(t *testing.T) {
	ws := &WorkspaceConfig{
		Environments: map[string]Environment{
			"dev": {
				Services: []Service{
					{Provider: "vscode", Folder: "/path"},
				},
			},
		},
	}
	err := Validate(ws)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Name")
}

func TestValidate_NameTooLong(t *testing.T) {
	name := ""
	for i := 0; i < 101; i++ {
		name += "a"
	}
	ws := &WorkspaceConfig{
		Name: name,
		Environments: map[string]Environment{
			"dev": {
				Services: []Service{
					{Provider: "vscode", Folder: "/path"},
				},
			},
		},
	}
	err := Validate(ws)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Name")
}

func TestValidate_EmptyEnvironments(t *testing.T) {
	ws := &WorkspaceConfig{
		Name: "test",
		Environments: map[string]Environment{},
	}
	err := Validate(ws)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Environments")
}

func TestValidate_EmptyServices(t *testing.T) {
	ws := &WorkspaceConfig{
		Name: "test",
		Environments: map[string]Environment{
			"dev": {
				Services: []Service{},
			},
		},
	}
	err := Validate(ws)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Services")
}

func TestValidate_InvalidProvider(t *testing.T) {
	ws := &WorkspaceConfig{
		Name: "test",
		Environments: map[string]Environment{
			"dev": {
				Services: []Service{
					{Provider: "unknown", Folder: "/path"},
				},
			},
		},
	}
	err := Validate(ws)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Provider")
}

func TestValidate_TerminalMissingName(t *testing.T) {
	ws := &WorkspaceConfig{
		Name: "test",
		Environments: map[string]Environment{
			"dev": {
				Services: []Service{
					{
						Provider: "vscode",
						Folder:   "/path",
						Terminals: []Terminal{
							{Directory: "/path/server"},
						},
					},
				},
			},
		},
	}
	err := Validate(ws)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Name")
}

func TestValidate_TerminalMissingDirectory(t *testing.T) {
	ws := &WorkspaceConfig{
		Name: "test",
		Environments: map[string]Environment{
			"dev": {
				Services: []Service{
					{
						Provider: "vscode",
						Folder:   "/path",
						Terminals: []Terminal{
							{Name: "Server"},
						},
					},
				},
			},
		},
	}
	err := Validate(ws)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Directory")
}

func TestValidate_EmptyScanPaths(t *testing.T) {
	cfg := &GlobalConfig{}
	err := Validate(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ScanPaths")
}

func TestValidate_AllValidProviders(t *testing.T) {
	providers := []string{"vscode", "dbeaver", "chrome", "docker", "command"}
	for _, p := range providers {
		ws := &WorkspaceConfig{
			Name: "test",
			Environments: map[string]Environment{
				"dev": {
					Services: []Service{
						{Provider: p, Folder: "/path"},
					},
				},
			},
		}
		err := Validate(ws)
		assert.NoError(t, err, "provider %s should be valid", p)
	}
}

func TestValidate_NilInput(t *testing.T) {
	err := Validate(nil)
	assert.Error(t, err)
}

type testCustomTag struct {
	Field string `validate:"eq=hello"`
}

func TestValidate_DefaultCase(t *testing.T) {
	v := testCustomTag{Field: "world"}
	err := Validate(v)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Field failed validation: eq")
}

func TestValidate_NonStructInput(t *testing.T) {
	err := Validate("not a struct")
	assert.Error(t, err)
}

func TestValidate_OneofTag(t *testing.T) {
	ws := &WorkspaceConfig{
		Name: "test",
		Environments: map[string]Environment{
			"dev": {
				Services: []Service{
					{Provider: "invalid", Folder: "/path"},
				},
			},
		},
	}
	err := Validate(ws)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be one of")
	assert.Contains(t, err.Error(), "Provider")
}

type testMaxTag struct {
	Name string `validate:"max=3"`
}

func TestValidate_MaxTag(t *testing.T) {
	v := testMaxTag{Name: "toolong"}
	err := Validate(v)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be at most")
}

type testMinTag struct {
	Name string `validate:"min=5"`
}

func TestValidate_MinTag(t *testing.T) {
	v := testMinTag{Name: "x"}
	err := Validate(v)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be at least")
}
