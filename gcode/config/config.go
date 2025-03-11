package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/xingty/rcode-go/pkg/utils"
)

const ENV_DEBUG = "GCODE_DEBUG"

var HOME, _ = os.UserHomeDir()

var GCODE_HOME = filepath.Join(HOME, ".gcode")
var GCCODE_CONFIG = filepath.Join(GCODE_HOME, "gcode")
var GCODE_KEY_FILE = filepath.Join(GCODE_HOME, "keyfile")
var RSSH_KEY_FILE = filepath.Join(HOME, ".rssh", "keyfile")

var SUPPORTED_IDE = utils.NewSet("code", "cursor", "windsurf")

func InitGCodeEnv() {
	if _, err := os.Stat(GCODE_HOME); os.IsNotExist(err) {
		println("GCODE_HOME not exist, creating...")
		os.Mkdir(GCODE_HOME, 0755)
	}

	if _, err := os.Stat(GCCODE_CONFIG); os.IsNotExist(err) {
		file, err := os.Create(GCCODE_CONFIG)
		if err != nil {
			panic(err)
		}
		file.Close()
	}

	if _, err := os.Stat(GCODE_KEY_FILE); os.IsNotExist(err) {
		file, err := os.Create(GCODE_KEY_FILE)
		if err != nil {
			panic(err)
		}

		file.Write([]byte(uuid.New().String()))
		file.Close()
	}

	initLogger()
}

func initLogger() {
	logDir := filepath.Join(GCODE_HOME, "logs")
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		os.MkdirAll(logDir, 0755)
	}
	logFilePath := filepath.Join(logDir, "ipc.log")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	log.SetOutput(logFile)
}
