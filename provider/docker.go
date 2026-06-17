package provider

import (
	"context"
	"errors"
	"os/exec"

	"github.com/anomalyco/nook/config"
	"github.com/anomalyco/nook/utils"
)

type DockerProvider struct{}

func (p *DockerProvider) Name() string {
	return "docker"
}

func (p *DockerProvider) Detect() (bool, error) {
	if _, err := exec.LookPath("docker"); err == nil {
		return true, nil
	}
	return false, nil
}

func (p *DockerProvider) Launch(ctx context.Context, svc config.Service, baseDir string, envVars map[string]string) error {
	if svc.File == "" {
		return errors.New("docker: file is required")
	}

	file := utils.ResolvePath(baseDir, svc.File)
	cmd := dockerCommand(ctx, file, svc.Profile)
	return cmd.Run()
}

func dockerCommand(ctx context.Context, file, profile string) *exec.Cmd {
	args := []string{"compose", "-f", file}
	if profile != "" {
		args = append(args, "--profile", profile)
	}
	args = append(args, "up", "-d")
	return execCommandContext(ctx, "docker", args...)
}

func init() {
	Register(&DockerProvider{})
}
