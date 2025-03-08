package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/xingty/rcode-go/gcode/code"
	"github.com/xingty/rcode-go/gcode/config"
)

var COMMANDS map[string]string = map[string]string{
	"gcode":     "code",
	"gcursor":   "cursor",
	"gwindsurf": "windsurf",
}

func main() {
	config.InitGCodeEnv()

	flag.Usage = func() {
		fmt.Println("Usage: rcode <host> <dir>")
		fmt.Println("just rcode 'file' like your VSCode 'code' .")
		fmt.Println("but you should config your ~/.ssh/config first")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
	}
	isLatest := flag.Bool("l", false, "if is_latest")
	shortcutName := flag.String("sn", "latest", "add shortcut name to this")
	openShortcutName := flag.String("os", "", "open shortcut name")
	flag.Parse()

	binName := os.Args[0]
	bin, ok := COMMANDS[binName]
	if !ok {
		fmt.Printf("unknown command: %s\n", binName)
		os.Exit(1)
	}
	binName = bin

	isRemote := code.IsRemote(binName)
	if isRemote {
		if len(os.Args) < 2 {
			flag.Usage()
			os.Exit(1)
		}

		dirName := os.Args[1]
		err := code.RunRemote(binName, dirName, code.MAX_IDLE_TIME)
		if err != nil {
			panic(err)
		}

		os.Exit(0)
	}

	if *isLatest {
		err := code.RunLatest(binName)
		if err != nil {
			fmt.Printf("failed to run %s: %s\n", binName, err.Error())
			os.Exit(1)
		}

		os.Exit(0)
	}

	if *openShortcutName != "" {
		err := code.RunShortcut(binName, *openShortcutName)
		if err != nil {
			fmt.Printf("failed to run %s: %s\n", binName, err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}

	if len(os.Args) < 3 {
		flag.Usage()
		os.Exit(1)
	}

	hostname := os.Args[1]
	if _, found := strings.CutPrefix(hostname, "-"); found {
		flag.Usage()
		os.Exit(1)
	}

	dirName := os.Args[2]
	if _, found := strings.CutPrefix(dirName, "-"); found {
		flag.Usage()
		os.Exit(1)
	}

	code.RunLocal(binName, hostname, dirName, *shortcutName)
}
