package code

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xingty/rcode-go/gcode/config"
	"github.com/xingty/rcode-go/pkg/utils/sshconf"
)

func RunLocal(
	binName string,
	hostname string,
	dirName string,
	shortcutName string) error {

	if binName == "zed" {
		return runZed(hostname, dirName)
	} else {
		return runVSCodeLikeIDE(binName, hostname, dirName, shortcutName)
	}
}

func OpenLocalPath(binName string, path string) error {
	if !config.SUPPORTED_IDE.Has(binName) {
		return fmt.Errorf("unsupported ide: %s", binName)
	}

	if binName == "zed" {
		return exec.Command("zed", path).Run()
	}
	return exec.Command(binName, path).Run()
}

func runZed(hostname string, dirName string) error {
	remoteURI := fmt.Sprintf("ssh://%s%s", hostname, dirName)
	return exec.Command("zed", remoteURI).Run()
}

func runVSCode(hostname string, dirName string) error {
	remoteURI := fmt.Sprintf("vscode-remote://ssh-remote+%s%s", hostname, dirName)
	return exec.Command("code", "--folder-uri", remoteURI).Run()
}

func expandDir(hostname string, dirName string, platform string) (string, error) {
	home, _ := os.UserHomeDir()

	cfgFile := filepath.Join(home, "/.ssh/config")
	config := sshconf.NewSSHConfig(cfgFile)
	host := config.GetHost(hostname)
	if host == nil {
		return "", errors.New("couldn't expand user home directory")
	}

	var homeDir string
	if platform == "linux" {
		homeDir = "/home/"
	} else if platform == "macos" {
		homeDir = "/Users/"
	} else {
		panic("unsupported platform: " + platform)
	}

	return homeDir + host.GetUser("root") + dirName[1:], nil
}

func runVSCodeLikeIDE(
	binName string,
	hostname string,
	dirName string,
	shortcutName string) error {

	if strings.HasPrefix(dirName, "~/") {
		var err error
		dirName, err = expandDir(hostname, dirName, "linux")
		if err != nil {
			return err
		}
	}

	AppendConfig(shortcutName, hostname, dirName)
	ReadAndMergeConfig("latest")
	return runVSCode(hostname, dirName)
}

func RunLatest(binName string) error {
	err := RunShortcut(binName, "latest")
	if err != nil {
		return errors.New("no latest config, please run 'gcode hostname dir' first")
	}

	return nil
}

func RunShortcut(binName string, shortcutName string) error {
	uri := ReadAndMergeConfig("latest")
	if len(uri) == 0 {
		return errors.New("no latest config, please run 'gcode hostname dir' first")
	}

	uris := strings.Split(uri, ",")
	if binName == "zed" {
		return runZed(uris[1], uris[2])
	} else {
		return runVSCode(uris[1], uris[2])
	}
}
