package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/xingty/rcode-go/pkg/utils"
)

const ENV_DEBUG = "GCODE_DEBUG"

var (
	HOME string

	GCODE_HOME     string
	GCCODE_CONFIG  string
	GCODE_KEY_FILE string
	RSSH_KEY_FILE  string
)

var SUPPORTED_IDE = utils.NewSet("code", "cursor", "windsurf", "zed", "trae")

var loggerMu sync.Mutex
var loggerInitialized bool

func init() {
	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		SetHomeDir(home)
	}
}

type SetupOptions struct {
	EnsureConfigFile bool
	EnsureKeyFile    bool
	InitLogger       bool
}

func SetHomeDir(home string) {
	HOME = home
	RSSH_KEY_FILE = filepath.Join(HOME, ".rssh", "keyfile")
	if GCODE_HOME == "" {
		SetGCodeHome(filepath.Join(HOME, ".gcode"))
	} else {
		updateDerivedPaths()
	}
}

func SetGCodeHome(path string) {
	GCODE_HOME = path
	updateDerivedPaths()
}

func updateDerivedPaths() {
	GCCODE_CONFIG = filepath.Join(GCODE_HOME, "gcode")
	GCODE_KEY_FILE = filepath.Join(GCODE_HOME, "keyfile")
}

func Setup(opts SetupOptions) (closer func() error, err error) {
	if HOME == "" {
		home, e := os.UserHomeDir()
		if e != nil {
			return nil, fmt.Errorf("failed to resolve user home dir: %w", e)
		}
		if home == "" {
			return nil, errors.New("failed to resolve user home dir: empty path")
		}
		SetHomeDir(home)
	}

	if GCODE_HOME == "" {
		return nil, errors.New("GCODE_HOME is empty")
	}
	updateDerivedPaths()

	if err := os.MkdirAll(GCODE_HOME, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create GCODE_HOME %q: %w", GCODE_HOME, err)
	}

	if opts.EnsureConfigFile {
		if err := ensureFileExists(GCCODE_CONFIG, 0o644); err != nil {
			return nil, err
		}
	}

	if opts.EnsureKeyFile {
		if err := ensureKeyFile(GCODE_KEY_FILE); err != nil {
			return nil, err
		}
	}

	if opts.InitLogger {
		loggerMu.Lock()
		defer loggerMu.Unlock()
		if loggerInitialized {
			return nil, nil
		}
		closer, err := initLogger()
		if err != nil {
			return nil, err
		}
		loggerInitialized = true
		return closer, nil
	}

	return nil, nil
}

func ensureFileExists(path string, perm os.FileMode) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to stat file %q: %w", path, err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, perm)
	if err != nil {
		return fmt.Errorf("failed to create file %q: %w", path, err)
	}
	return f.Close()
}

func ensureKeyFile(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to stat key file %q: %w", path, err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err == nil {
		if _, werr := f.Write([]byte(uuid.NewString())); werr != nil {
			_ = f.Close()
			return fmt.Errorf("failed to write key file %q: %w", path, werr)
		}
		return f.Close()
	}
	if errors.Is(err, os.ErrExist) {
		return nil
	}
	return fmt.Errorf("failed to create key file %q: %w", path, err)
}

func initLogger() (func() error, error) {
	logDir := filepath.Join(GCODE_HOME, "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create log dir %q: %w", logDir, err)
	}
	logFilePath := filepath.Join(logDir, "ipc.log")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %q: %w", logFilePath, err)
	}
	log.SetOutput(logFile)
	return logFile.Close, nil
}
