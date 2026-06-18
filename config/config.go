package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"gopkg.in/yaml.v3"
)

type GlobalConfig struct {
	ScanPaths []string `yaml:"scan_paths" validate:"required,min=1"`
}

func configDirPath() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = xdg.ConfigHome
	}
	return filepath.Join(configHome, "nook")
}

func configFilePath() string {
	return filepath.Join(configDirPath(), "config.yaml")
}

func LoadGlobalConfig() (*GlobalConfig, error) {
	dir := configDirPath()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	cfgPath := configFilePath()
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &GlobalConfig{
				ScanPaths: []string{"~/.nook/workspaces"},
			}, nil
		}
		return nil, err
	}

	cfg := &GlobalConfig{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	for i, p := range cfg.ScanPaths {
		cfg.ScanPaths[i] = expandTilde(p)
	}

	if len(cfg.ScanPaths) == 0 {
		cfg.ScanPaths = []string{expandTilde("~/.nook/workspaces")}
	}

	return cfg, nil
}

func SaveGlobalConfig(cfg *GlobalConfig) error {
	dir := configDirPath()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	cfgPath := configFilePath()
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(cfgPath, data, 0644)
}

func expandTilde(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
