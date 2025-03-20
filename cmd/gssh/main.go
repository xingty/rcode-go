package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/xingty/rcode-go/gcode/config"
	"github.com/xingty/rcode-go/gcode/ssh"
)

var version = "0.0.10"

func main() {
	for _, arg := range os.Args[1:] {
		if arg == "-R" || arg == "-T" {
			fmt.Printf("Error: %s is not allowed\n", arg)
			os.Exit(1)
		}
	}

	var host string
	var port int
	var v bool

	flag.StringVar(&host, "host", "127.0.0.1", "IPC server host")
	flag.IntVar(&port, "port", 7532, "IPC server port")
	flag.BoolVar(&v, "v", false, "Show version")
	flag.Parse()

	if v {
		fmt.Printf("gssh version: %s %s/%s\n", version, runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	config.InitGCodeEnv()
	ssh.Run(host, port, flag.Args())
}
