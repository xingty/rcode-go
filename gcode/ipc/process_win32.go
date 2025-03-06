//go:build windows
// +build windows

package ipc

import (
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

func StartSSHClient(args []string) error {
	cmd := exec.Command("ssh", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
