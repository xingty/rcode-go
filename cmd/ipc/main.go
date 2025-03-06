package main

import (
	"flag"

	"github.com/xingty/rcode-go/gcode/config"
	"github.com/xingty/rcode-go/gcode/ipc"
)

func main() {
	var host string
	var port int
	var maxIdleTime int

	flag.StringVar(&host, "host", "127.0.0.1", "IPC server host")
	flag.IntVar(&port, "port", 7532, "IPC server port")
	flag.IntVar(&maxIdleTime, "max-idle", 600, "Max idle time in seconds")
	flag.Parse()

	config.InitGCodeEnv()
	server := ipc.NewIPCServerSocket(maxIdleTime)
	server.Start(host, port)
}
