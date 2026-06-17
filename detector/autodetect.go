package detector

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/lorenzo-vecchio/nook/config"
)

type WorkspaceInfo struct {
	Name        string
	Description string
	Path        string
	EnvNames    []string
}

type DetectionResult struct {
	InCWD     []WorkspaceInfo
	InSubdirs []WorkspaceInfo
	HasNew    bool
}

func ScanCurrentDir() (*DetectionResult, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	result := &DetectionResult{}

	cwdInfo, err := scanSingleDir(cwd)
	if err != nil {
		return nil, err
	}
	if cwdInfo != nil {
		result.InCWD = append(result.InCWD, *cwdInfo)
		result.HasNew = true
	}

	entries, err := os.ReadDir(cwd)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", cwd, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			subPath := filepath.Join(cwd, entry.Name())
			info, err := scanSingleDir(subPath)
			if err != nil {
				return nil, err
			}
			if info != nil {
				result.InSubdirs = append(result.InSubdirs, *info)
				result.HasNew = true
			}
		}
	}

	return result, nil
}

func ScanPath(path string) (map[string]*WorkspaceInfo, error) {
	workspaces := make(map[string]*WorkspaceInfo)

	info, err := scanSingleDir(path)
	if err != nil {
		return nil, err
	}
	if info != nil {
		workspaces[info.Name] = info
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", path, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			subPath := filepath.Join(path, entry.Name())
			info, err := scanSingleDir(subPath)
			if err != nil {
				return nil, err
			}
			if info != nil {
				workspaces[info.Name] = info
			}
		}
	}

	return workspaces, nil
}

func IsTrusted(dir string, scanPaths []string) bool {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return false
	}
	absDir = filepath.Clean(absDir)

	for _, sp := range scanPaths {
		absSP, err := filepath.Abs(sp)
		if err != nil {
			continue
		}
		absSP = filepath.Clean(absSP)

		if absDir == absSP {
			return true
		}

		rel, err := filepath.Rel(absSP, absDir)
		if err != nil {
			continue
		}
		if !filepath.IsLocal(rel) {
			continue
		}
		return true
	}

	return false
}

func TrustPath(path string) error {
	cfg, err := config.LoadGlobalConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	if !slices.Contains(cfg.ScanPaths, absPath) {
		cfg.ScanPaths = append(cfg.ScanPaths, absPath)
	}

	return config.SaveGlobalConfig(cfg)
}

func scanSingleDir(dir string) (*WorkspaceInfo, error) {
	yamlPath := filepath.Join(dir, "workspace.yaml")
	info, err := os.Stat(yamlPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to stat %s: %w", yamlPath, err)
	}
	if info.IsDir() {
		return nil, nil
	}

	ws, err := config.LoadWorkspace(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", yamlPath, err)
	}

	envNames := make([]string, 0, len(ws.Environments))
	for name := range ws.Environments {
		envNames = append(envNames, name)
	}
	if envNames == nil {
		envNames = []string{}
	}

	return &WorkspaceInfo{
		Name:        ws.Name,
		Description: ws.Description,
		Path:        dir,
		EnvNames:    envNames,
	}, nil
}
