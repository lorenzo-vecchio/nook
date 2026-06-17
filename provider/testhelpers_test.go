package provider

import (
	"context"
	"os/exec"
	"runtime"
)

func testCmd(ctx context.Context, arg ...string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, "cmd", "/c", "echo")
	}
	return exec.CommandContext(ctx, "/bin/echo", arg...)
}
