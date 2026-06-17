package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type WorkspaceConfig struct {
	Name         string                 `yaml:"name" validate:"required,min=1,max=100"`
	Description  string                 `yaml:"description"`
	Environments map[string]Environment `yaml:"environments" validate:"required,min=1,dive"`
}

type Environment struct {
	EnvFile  string    `yaml:"env_file"`
	Services []Service `yaml:"services" validate:"required,min=1,dive"`
}

type Service struct {
	Provider   string     `yaml:"provider" validate:"required,oneof=vscode dbeaver chrome docker command"`
	Folder     string     `yaml:"folder,omitempty"`
	Terminals  []Terminal `yaml:"terminals,omitempty" validate:"omitempty,dive"`
	Connection string     `yaml:"connection,omitempty"`
	URLs       []string   `yaml:"urls,omitempty"`
	File       string     `yaml:"file,omitempty"`
	Profile    string     `yaml:"profile,omitempty"`
	Cmd        string     `yaml:"cmd,omitempty"`
	Cwd        string     `yaml:"cwd,omitempty"`
}

type Terminal struct {
	Name      string `yaml:"name" validate:"required"`
	Directory string `yaml:"directory" validate:"required"`
	Command   string `yaml:"command"`
}

func LoadWorkspace(path string) (*WorkspaceConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	ws := &WorkspaceConfig{}
	if err := yaml.Unmarshal(data, ws); err != nil {
		return nil, err
	}

	return ws, nil
}

func SaveWorkspace(cfg *WorkspaceConfig, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
