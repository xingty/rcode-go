package config

import (
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/xingty/rcode-go/pkg/utils"
)

var GCODE_HOME = filepath.Join(os.Getenv("HOME"), ".gcode")
var GCCODE_CONFIG = filepath.Join(GCODE_HOME, "gcode")
var GCODE_KEY_FILE = filepath.Join(GCODE_HOME, "keyfile")

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
}
