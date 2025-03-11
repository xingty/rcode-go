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

func StartSSHClient(args []string) error {
	path, err := exec.LookPath("ssh")
	if err != nil {
		fmt.Println(err)
		return err
	}

	newArgs := append([]string{"ssh"}, args...)
	return syscall.Exec(path, newArgs, os.Environ())
}
