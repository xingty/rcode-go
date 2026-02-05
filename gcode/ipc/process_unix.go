//go:build !windows
// +build !windows

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
		Setsid: true,
	}

	return cmd.Start()
}

func StartSSHClient(args []string) int {
	path, err := exec.LookPath("ssh")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 255
	}

	newArgs := append([]string{"ssh"}, args...)
	if err := syscall.Exec(path, newArgs, os.Environ()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 255
	}

	// unreachable on success, but keeps the compiler happy.
	return 0
}
