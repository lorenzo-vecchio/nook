package provider

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/lorenzo-vecchio/nook/config"
	"github.com/lorenzo-vecchio/nook/utils"
)

type ChromeProvider struct {
	detectedPath string
}

func (p *ChromeProvider) Name() string {
	return "chrome"
}

func (p *ChromeProvider) Detect() (bool, error) {
	names := chromeBinaryNames()
	for _, name := range names {
		path, err := exec.LookPath(name)
		if err == nil {
			p.detectedPath = path
			return true, nil
		}
	}

	for _, path := range chromePaths() {
		if _, err := os.Stat(path); err == nil {
			p.detectedPath = path
			return true, nil
		}
	}

	return false, nil
}

func (p *ChromeProvider) Launch(ctx context.Context, svc config.Service, baseDir string, envVars map[string]string) error {
	if p.detectedPath == "" {
		ok, err := p.Detect()
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("chrome not found")
		}
	}

	if len(svc.URLs) == 0 {
		return nil
	}

	var normalized []string
	for _, url := range svc.URLs {
		normalized = append(normalized, normalizeURL(url))
	}

	cmd := chromeCommand(ctx, p.detectedPath, normalized)
	return cmd.Start()
}

func chromeBinaryNames() []string {
	if utils.IsWindows() {
		return []string{"chrome.exe"}
	}
	return []string{"google-chrome", "google-chrome-stable", "chromium-browser", "chromium"}
}

func chromePaths() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
		}
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		return []string{
			filepath.Join(os.Getenv("ProgramFiles"), "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(localAppData, "Google", "Chrome", "Application", "chrome.exe"),
		}
	default:
		return []string{
			"/usr/bin/google-chrome",
			"/usr/bin/google-chrome-stable",
			"/usr/bin/chromium-browser",
			"/usr/bin/chromium",
		}
	}
}

func normalizeURL(url string) string {
	if !strings.Contains(url, "://") {
		return "http://" + url
	}
	return url
}

func chromeCommand(ctx context.Context, chromePath string, urls []string) *exec.Cmd {
	if utils.IsWindows() {
		args := []string{"/c", "start", "chrome"}
		args = append(args, urls...)
		return execCommandContext(ctx, "cmd", args...)
	}
	args := []string{}
	for _, url := range urls {
		args = append(args, "--new-tab", url)
	}
	return execCommandContext(ctx, chromePath, args...)
}

func init() {
	Register(&ChromeProvider{})
}
