package provider

import (
	"context"

	"github.com/anomalyco/nook/config"
)

type Provider interface {
	Name() string
	Detect() (bool, error)
	Launch(ctx context.Context, service config.Service, baseDir string, envVars map[string]string) error
}
