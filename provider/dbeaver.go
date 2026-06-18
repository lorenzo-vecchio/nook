package provider

import (
	"context"
	"os"
	"os/exec"
	"runtime"

	"github.com/lorenzo-vecchio/nook/config"
	"github.com/lorenzo-vecchio/nook/utils"
)

type DBeaverProvider struct{}

func (p *DBeaverProvider) Name() string {
	return "dbeaver"
}

func (p *DBeaverProvider) Detect() (bool, error) {
	if _, err := exec.LookPath("dbeaver"); err == nil {
		return true, nil
	}
	if utils.IsWindows() {
		if _, err := exec.LookPath("dbeaver-cli"); err == nil {
			return true, nil
		}
	}

	for _, path := range dbeaverCommonPaths() {
		if _, err := os.Stat(path); err == nil {
			return true, nil
		}
	}

	return false, nil
}

func (p *DBeaverProvider) Launch(ctx context.Context, svc config.Service, baseDir string, envVars map[string]string) error {
	dbeaverPath := p.findDBeaverPath()
	cmd := execCommandContext(ctx, dbeaverPath, "-con", svc.Connection)
	return cmd.Start()
}

func (p *DBeaverProvider) findDBeaverPath() string {
	if path, err := exec.LookPath("dbeaver"); err == nil {
		return path
	}
	if utils.IsWindows() {
		if path, err := exec.LookPath("dbeaver-cli"); err == nil {
			return path
		}
	}

	for _, path := range dbeaverCommonPaths() {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "dbeaver"
}

func dbeaverCommonPaths() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{
			"/Applications/DBeaver.app/Contents/MacOS/dbeaver",
		}
	case "windows":
		return []string{
			`C:\Program Files\DBeaver\dbeaver-cli.exe`,
		}
	default:
		return nil
	}
}

func init() {
	Register(&DBeaverProvider{})
}
