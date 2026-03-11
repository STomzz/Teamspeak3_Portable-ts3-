//go:build !windows

package main

import (
	"fmt"
	"os"
	"os/exec"
)

func showErrorDialog(_, _ string) {}

func launchClient(exePath string, args []string, workDir string, env []string) error {
	cmd := exec.Command(exePath, args...)
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = env
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("launch process: %w", err)
	}
	return nil
}
