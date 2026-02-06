package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/xingty/rcode-go/cmd/internal/gsshcli"
	"github.com/xingty/rcode-go/gcode/config"
	"github.com/xingty/rcode-go/gcode/ssh"
)

var version = "0.0.10"

func main() {
	if handled, exitCode := maybeHandleInternalPrepare(os.Args[1:]); handled {
		os.Exit(exitCode)
	}

	opts, sshArgs, err := gsshcli.ParseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	if opts.Help {
		fmt.Fprintln(os.Stdout, gsshcli.Usage("gssh"))
		return
	}
	if opts.Version {
		fmt.Printf("gssh version: %s %s/%s\n", version, runtime.GOOS, runtime.GOARCH)
		return
	}

	closer, err := config.Setup(config.SetupOptions{
		EnsureKeyFile: true,
		InitLogger:    true,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	if closer != nil {
		defer closer()
	}

	os.Exit(ssh.Run(opts.Host, opts.Port, sshArgs))
}
