package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
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

		fmt.Printf("Usage: [%s] <host> <dir> [options]\n", keys)
		fmt.Println("just gcode 'file' like your VSCode 'code' .")
		fmt.Println("but you should config your ~/.ssh/config first")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
	}

	args = args[1:]
	commands := make([]string, 0)
	for index, arg := range args {
		if strings.HasPrefix(arg, "-") {
			arg = arg[index:]
			break
		}

		commands = append(commands, arg)
	}

	isRemote, err := code.IsRemote(binName)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
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

	isLatest := flag.Bool("l", false, "if is_latest")
	shortcutName := flag.String("sn", "latest", "open shortcut name")
	openShortcut := flag.String("os", "", "open shortcut")
	flag.CommandLine.Parse(args)

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
