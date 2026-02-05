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
)

var IDE_BY_INVOKED = map[string]string{
	"gcode":     "code",
	"gcursor":   "cursor",
	"gwindsurf": "windsurf",
	"gzed":      "zed",
	"gtrae":     "trae",
}

var VALID_IDES = map[string]struct{}{
	"code":     {},
	"cursor":   {},
	"windsurf": {},
	"zed":      {},
	"trae":     {},
}

var version = "0.0.10"

func main() {
	// config.InitGCodeEnv()
	invoked := filepath.Base(os.Args[0])
	invoked = strings.TrimSuffix(invoked, ".exe")

	defaultIDE, ok := IDE_BY_INVOKED[invoked]
	if !ok {
		defaultIDE = IDE_BY_INVOKED["gcode"]
	}

	flag.Usage = func() {
		keys := strings.Join(lo.Keys(IDE_BY_INVOKED), " | ")

		fmt.Println("Usage:")
		fmt.Printf("Run on local:  [%s] [--ide <ide>] <host> <dir> [options]\n", keys)
		fmt.Printf("Run on remote: [%s] [--ide <ide>] <dir>\n", keys)
		fmt.Println("Just gcode 'file' like your VSCode 'code'.")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
	}

	v := flag.Bool("v", false, "Show version")
	ide := flag.String("ide", "", "IDE to use: code | cursor | windsurf | zed | trae")
	isLatest := flag.Bool("l", false, "if is_latest")
	shortcutName := flag.String("sn", "latest", "open shortcut name")
	openShortcut := flag.String("os", "", "open shortcut")
	flag.CommandLine.Parse(os.Args[1:])

	if *v {
		fmt.Printf("gcode version: %s %s/%s\n", version, runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	commands := flag.Args()
	if invoked == "gcode" && len(commands) > 0 {
		if legacyIDE, ok := IDE_BY_INVOKED[commands[0]]; ok {
			defaultIDE = legacyIDE
			commands = commands[1:]
		}
	}

	ideName := defaultIDE
	if *ide != "" {
		ideName = *ide
	}
	if _, ok := VALID_IDES[ideName]; !ok {
		fmt.Printf("unsupported ide: %s\n", ideName)
		os.Exit(1)
	}

	isRemote, err := code.IsRemote(ideName)
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
		err := code.RunRemote(ideName, dirName, code.MAX_IDLE_TIME)
		if err != nil {
			fmt.Printf("failed to run %s: %s\n", ideName, err.Error())
			os.Exit(1)
		}

		os.Exit(0)
	}

	if len(commands) == 1 {
		path := commands[0]
		if strings.HasPrefix(path, "~/") {
			home, _ := os.UserHomeDir()
			path = filepath.Join(home, path[2:])
		}
		path, _ = filepath.Abs(path)

		err := code.OpenLocalPath(ideName, path)
		if err != nil {
			fmt.Printf("failed to open %s: %s\n", ideName, err.Error())
			os.Exit(1)
		}

		os.Exit(0)
	}

	if len(commands) >= 2 {
		hostname := commands[0]
		dirName := commands[1]

		err := code.RunLocal(ideName, hostname, dirName, *shortcutName)
		if err != nil {
			fmt.Printf("failed to run %s: %s\n", ideName, err.Error())
			os.Exit(1)
		}

		os.Exit(0)
	}

	if *isLatest {
		err := code.RunLatest(ideName)
		if err != nil {
			fmt.Printf("failed to run %s: %s\n", ideName, err.Error())
			os.Exit(1)
		}

		os.Exit(0)
	}

	if *openShortcut != "" {
		err := code.RunShortcut(ideName, *shortcutName)
		if err != nil {
			fmt.Printf("failed to run %s: %s\n", ideName, err.Error())
			os.Exit(1)
		}

		os.Exit(0)
	}

	flag.Usage()
	os.Exit(1)
}
