package source

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Runner executes external commands for source refresh operations.
type Runner interface {
	Run(ctx context.Context, workdir string, name string, args ...string) error
}

// ExecRunner executes commands through the local operating system process runner.
type ExecRunner struct{}

// Run executes one command and returns any combined output as part of the error path.
func (ExecRunner) Run(ctx context.Context, workdir string, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	if workdir != "" {
		cmd.Dir = workdir
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			return fmt.Errorf("run %s %s: %w", name, strings.Join(args, " "), err)
		}

		return fmt.Errorf("run %s %s: %w: %s", name, strings.Join(args, " "), err, message)
	}

	return nil
}
