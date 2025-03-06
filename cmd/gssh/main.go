package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/xingty/rcode-go/gcode/config"
	"github.com/xingty/rcode-go/gcode/ssh"
)

func main() {
	for _, arg := range os.Args[1:] {
		if arg == "-R" || arg == "-T" {
			fmt.Printf("Error: %s is not allowed\n", arg)
			os.Exit(1)
		}
	}

	var host string
	var port int

	flag.StringVar(&host, "host", "127.0.0.1", "IPC server host")
	flag.IntVar(&port, "port", 7532, "IPC server port")
	flag.Parse()

	config.InitGCodeEnv()
	ssh.Run(host, port, flag.Args())
}
