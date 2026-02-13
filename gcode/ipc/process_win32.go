//go:build windows
// +build windows

package ipc

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

const detachedProcess uint32 = 0x00000008 // DETACHED_PROCESS (winbase.h)
const createNoWindow uint32 = 0x08000000  // CREATE_NO_WINDOW (winbase.h)

func StartIPCServer(binName string, args []string) error {
	exe := binName
	if !strings.HasSuffix(strings.ToLower(exe), ".exe") {
		exe += ".exe"
	}

	if self, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(self), exe)
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
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | detachedProcess,
	}
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

func configureNoConsoleWindow(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.HideWindow = true
	cmd.SysProcAttr.CreationFlags |= createNoWindow
}
