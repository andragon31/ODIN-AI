package guardian

import (
	"context"
	"os/exec"
)

// execCommandContext is a wrapper around exec.CommandContext
// This allows for easier testing and mocking
func execCommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd
}

// execCommand is a wrapper around exec.Command (without context)
func execCommand(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	return cmd
}
