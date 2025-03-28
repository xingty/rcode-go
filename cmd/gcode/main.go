package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/samber/lo"
	"github.com/xingty/rcode-go/gcode/code"
	"github.com/xingty/rcode-go/gcode/config"
)

var COMMANDS map[string]string = map[string]string{
	"gcode":     "code",
	"gcursor":   "cursor",
	"gwindsurf": "windsurf",
	"gtrae":     "trae",
}

var version = "0.0.10"

func main() {
	config.InitGCodeEnv()
	args := os.Args[1:]
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	binName, ok := COMMANDS[args[0]]
	if !ok {
		fmt.Printf("unknown command: %s\n", binName)
		os.Exit(1)
	}

	flag.Usage = func() {
		keys := strings.Join(lo.Keys(COMMANDS), " | ")

		fmt.Println("Usage:")
		fmt.Printf("Run on local:  [%s] <host> <dir> [options]\n", keys)
		fmt.Printf("Run on remote: [%s] <dir> \n", keys)
		fmt.Println("Just gcode 'file' like your VSCode 'code' .")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
	}

	args = args[1:]
	commands := make([]string, 0)
	for index, arg := range args {
		if strings.HasPrefix(arg, "-") {
			args = args[index:]
			break
		}

		commands = append(commands, arg)
	}

	isRemote, err := code.IsRemote(binName)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	v := flag.Bool("v", false, "Show version")
	isLatest := flag.Bool("l", false, "if is_latest")
	shortcutName := flag.String("sn", "latest", "open shortcut name")
	openShortcut := flag.String("os", "", "open shortcut")
	flag.CommandLine.Parse(args)

	if *v {
		fmt.Printf("gcode version: %s %s/%s\n", version, runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	if isRemote {
		if len(commands) == 0 {
			flag.Usage()
			os.Exit(1)
		}

		dirName := commands[0]
		dirName, _ = filepath.Abs(dirName)
		err := code.RunRemote(binName, dirName, code.MAX_IDLE_TIME)
		if err != nil {
			fmt.Printf("failed to run %s: %s\n", binName, err.Error())
			os.Exit(1)
		}

		os.Exit(0)
	}

	if len(commands) >= 2 {
		hostname := commands[0]
		dirName := commands[1]

		err := code.RunLocal(binName, hostname, dirName, *shortcutName)
		if err != nil {
			fmt.Printf("failed to run %s: %s\n", binName, err.Error())
			os.Exit(1)
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

	if *openShortcut != "" {
		err := code.RunShortcut(binName, *shortcutName)
		if err != nil {
			fmt.Printf("failed to run %s: %s\n", binName, err.Error())
			os.Exit(1)
		}

		os.Exit(0)
	}

	flag.Usage()
	os.Exit(1)
}
