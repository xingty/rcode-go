package code

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xingty/rcode-go/gcode/config"
)

type GCodeConfig struct {
	Host string
	Dir  string
}

func AppendConfig(name, host, dir string) {
	file := filepath.Join(config.GCODE_HOME, "gcode")
	fs, _ := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer fs.Close()
	fs.WriteString(fmt.Sprintf("%s,%s,%s\n", name, host, dir))
}

func WriteConfig(content string) {
	file := filepath.Join(config.GCODE_HOME, "gcode")
	fs, err := os.Create(file)
	if err != nil {
		panic(fmt.Sprintf("failed to open config file: %s, error: %v", file, err))
	}
	defer fs.Close()
	fs.WriteString(content)
}

func ParseOldURI(uri string) (string, string) {
	uri, _ = strings.CutPrefix(uri, "vscode-remote://ssh-remote+")
	index := strings.Index(uri, "/")

	return uri[:index], uri[index:]
}

func ReadAndMergeConfig(name string) string {
	confFile := filepath.Join(config.GCODE_HOME, "gcode")
	content, err := os.ReadFile(confFile)
	if err != nil {
		panic(err)
	}

	configList := []string{}
	values := make(map[string]int)
	lines := strings.Split(string(content), "\n")
	var remoteURI string
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if len(line) == 0 {
			continue
		}

		segs := strings.Split(lines[i], ",")
		key := strings.TrimSpace(segs[0])
		lastSeg := strings.TrimSpace(segs[len(segs)-1])

		_, ok := values[key]
		if ok {
			continue
		}

		var hostname, dirName string
		if strings.HasPrefix(lastSeg, "vscode-remote://ssh-remote+") {
			hostname, dirName = ParseOldURI(lastSeg)
			remoteURI = fmt.Sprintf("%s,%s,%s", key, hostname, dirName)
		} else {
			remoteURI = line
		}

		values[key] = len(configList)
		configList = append(configList, remoteURI)
	}

	newContent := strings.Join(configList, "\n")
	if len(strings.TrimSpace(newContent)) > 0 {
		WriteConfig(newContent + "\n")
	}

	index, ok := values[name]
	if !ok {
		return ""
	}

	return configList[index]
}
