package utils

import (
	"os"
	"runtime"
)

func IsMacOS() bool {
	return runtime.GOOS == "darwin"
}

func IsLinux() bool {
	return runtime.GOOS == "linux"
}

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func DefaultEditor() string {
	editor := os.Getenv("EDITOR")
	if editor != "" {
		return editor
	}
	if IsWindows() {
		return "notepad"
	}
	return "vim"
}

func ShellCommand(cmd string) (string, []string) {
	if IsWindows() {
		return "cmd", []string{"/c", cmd}
	}
	return "sh", []string{"-c", cmd}
}
