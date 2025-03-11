package main

import (
	"fmt"
	"os"
	"path/filepath"

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

	usage := func() {
		fmt.Println("Usage: rcode <host> <dir> [options]")
		fmt.Println("just rcode 'file' like your VSCode 'code' .")
		fmt.Println("but you should config your ~/.ssh/config first")
		fmt.Println("\nOptions:")
		fmt.Println("  -l    if is_latest")
		fmt.Println("  -sn   string")
		fmt.Println("        open shortcut name")
		fmt.Println("  -os   string")
		fmt.Println("        open shortcut name")
	}

	args := os.Args[1:]
	if len(args) == 0 {
		usage()
		os.Exit(1)
	}

	binName, ok := COMMANDS[args[0]]
	if !ok {
		fmt.Printf("unknown command: %s\n", binName)
		os.Exit(1)
	}

	isRemote, err := code.IsRemote(binName)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if isRemote {
		if len(args) < 2 {
			usage()
			os.Exit(1)
		}

		dirName := args[1]
		dirName, _ = filepath.Abs(dirName)
		err := code.RunRemote(binName, dirName, code.MAX_IDLE_TIME)
		if err != nil {
			panic(err)
		}

		os.Exit(0)
	}

	args = args[1:]
	if len(args) == 0 {
		usage()
		os.Exit(1)
	}

	isLatest := false
	openShortcut := false
	shortcutName := "latest"
	commands := make([]string, 0)
	for i, arg := range args {
		if arg == "-l" {
			isLatest = true
		} else if arg == "-os" {
			openShortcut = true
			if len(args) > i+1 {
				shortcutName = args[i+1]
				i += 1
			}
		} else if arg == "-sn" {
			if len(args) > i+1 {
				shortcutName = args[i+1]
				i += 1
			}
		} else {
			commands = append(commands, arg)
		}
	}

	if len(commands) >= 2 {
		hostname := commands[0]
		dirName := commands[1]

		err := code.RunLocal(binName, hostname, dirName, shortcutName)
		if err != nil {
			fmt.Printf("failed to run %s: %s\n", binName, err.Error())
			os.Exit(1)
		}

		os.Exit(0)
	}

	if isLatest {
		err := code.RunLatest(binName)
		if err != nil {
			fmt.Printf("failed to run %s: %s\n", binName, err.Error())
			os.Exit(1)
		}

		os.Exit(0)
	}

	if openShortcut {
		err := code.RunShortcut(binName, shortcutName)
		if err != nil {
			fmt.Printf("failed to run %s: %s\n", binName, err.Error())
			os.Exit(1)
		}

		os.Exit(0)
	}

	usage()
	os.Exit(1)
}
