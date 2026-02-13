//go:build !windows
// +build !windows

package ipc

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func StartIPCServer(binName string, args []string) error {
	exe := binName
	if self, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(self), binName)
		if st, statErr := os.Stat(candidate); statErr == nil && !st.IsDir() {
			exe = candidate
		}
	}

	_, err := StartDetached(exe, args)
	return err
}

func StartDetached(exe string, args []string) (*os.Process, error) {
	cmd := exec.Command(exe, args...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd.Process, nil
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

func configureNoConsoleWindow(_ *exec.Cmd) {}
