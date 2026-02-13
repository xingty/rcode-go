package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/xingty/rcode-go/gcode/config"
	"github.com/xingty/rcode-go/gcode/ipc"
)

var version = "0.0.10"

func main() {
	var host string
	var port int
	var maxIdleTime int
	var v bool
	var daemon bool

	flag.StringVar(&host, "host", "127.0.0.1", "IPC server host")
	flag.IntVar(&port, "port", 7532, "IPC server port")
	flag.IntVar(&maxIdleTime, "max-idle", 600, "Max idle time in seconds")
	flag.BoolVar(&v, "v", false, "Show version")
	flag.BoolVar(&daemon, "d", false, "Run as daemon (detach)")
	flag.Parse()

	if v {
		fmt.Printf("gssh-ipc version: %s %s/%s\n", version, runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	if daemon {
		args := make([]string, 0, len(os.Args)-1)
		for _, a := range os.Args[1:] {
			if a == "-d" || a == "-d=true" {
				continue
			}
			args = append(args, a)
		}

		exe, err := os.Executable()
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		proc, err := ipc.StartDetached(exe, args)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		fmt.Fprintf(os.Stdout, "gssh-ipc started (pid %d)\n", proc.Pid)
		os.Exit(0)
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

	server := ipc.NewIPCServerSocket(maxIdleTime)
	server.Start(host, port)
}
