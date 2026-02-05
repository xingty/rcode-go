//go:build windows
// +build windows

package ipc

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func StartIPCServer(binName string, args []string) error {
	cmd := exec.Command(binName, args...)

	cmd.Stdout = nil
	cmd.Stderr = nil

	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	return cmd.Start()
}

func StartSSHClient(args []string) int {
	path, err := exec.LookPath("ssh")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 255
	}

	cmd := exec.Command(path, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}

		fmt.Fprintln(os.Stderr, err)
		return 255
	}

	return 0
}
