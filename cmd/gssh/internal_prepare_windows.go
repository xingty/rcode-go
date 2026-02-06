//go:build windows

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/xingty/rcode-go/cmd/internal/gsshcli"
	"github.com/xingty/rcode-go/gcode/ssh"
)

type preparedCommand struct {
	Program string   `json:"program"`
	Args    []string `json:"args"`
}

func maybeHandleInternalPrepare(argv []string) (handled bool, exitCode int) {
	if len(argv) == 0 || argv[0] != "__prepare" {
		return false, 0
	}

	opts, sshArgs, err := gsshcli.ParseArgs(argv[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return true, 2
	}
	if opts.Help || opts.Version {
		fmt.Fprintln(os.Stderr, "__prepare is an internal command and does not support --help/--version")
		return true, 2
	}

	program, err := exec.LookPath("ssh")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return true, 255
	}

	ownerPID := int32(os.Getppid())
	if ownerPID <= 0 {
		ownerPID = int32(os.Getpid())
	}

	newArgs, err := ssh.BuildSSHArgs(opts.Host, opts.Port, ownerPID, sshArgs)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return true, 255
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(preparedCommand{Program: program, Args: newArgs}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return true, 255
	}

	return true, 0
}
