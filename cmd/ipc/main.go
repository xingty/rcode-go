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

	flag.StringVar(&host, "host", "127.0.0.1", "IPC server host")
	flag.IntVar(&port, "port", 7532, "IPC server port")
	flag.IntVar(&maxIdleTime, "max-idle", 600, "Max idle time in seconds")
	flag.BoolVar(&v, "v", false, "Show version")
	flag.Parse()

	if v {
		fmt.Printf("gssh-ipc version: %s %s/%s\n", version, runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	config.InitGCodeEnv()
	server := ipc.NewIPCServerSocket(maxIdleTime)
	server.Start(host, port)
}
