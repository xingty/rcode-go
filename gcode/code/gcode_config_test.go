package code

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xingty/rcode-go/gcode/config"
)

func setupTestEnv(t *testing.T) (cleanupFunc func()) {
	t.Helper()

	tempDir, err := ioutil.TempDir("", "test_gcode")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	oldGCODE_HOME := config.GCODE_HOME
	config.SetGCodeHome(tempDir)

	return func() {
		os.RemoveAll(tempDir)
		config.SetGCodeHome(oldGCODE_HOME)
	}
}

func TestAppendConfig(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	AppendConfig("latest", "debian", "/home/bigcat/gcode")

	configFile := filepath.Join(config.GCODE_HOME, "gcode")
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	expected := "latest,debian,/home/bigcat/gcode\n"
	if string(content) != expected {
		t.Errorf("AppendConfig() = %q, want %q", string(content), expected)
	}
}

// TestWriteConfig tests the WriteConfig function for proper file write operation.
func TestWriteConfig(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	histories := []string{
		"latest,debian,/home/bigcat/gcode",
		"ubuntu,ubuntu,/home/ubuntu/gcode",
	}

	testContent := strings.Join(histories, "\n") + "\n"
	WriteConfig(testContent)

	configFile := filepath.Join(config.GCODE_HOME, "gcode")
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("WriteConfig() = %q, want %q", strings.TrimSpace(string(content)), testContent)
	}
}

func TestReadAndMergeConfig(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	testConfigs := []string{
		"latest,debian,/home/bigcat/gcode",
		"old,vscode-remote://ssh-remote+debian/home/darbula/apps/monitor",
		"ubuntu,ubuntu,/home/ubuntu/gcode",
		"latest,ubuntu,/home/bigcat/gcode",
	}

	testContent := strings.Join(testConfigs, "\n") + "\n"
	configFile := filepath.Join(config.GCODE_HOME, "gcode")
	if err := os.WriteFile(configFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	latest := strings.Split(ReadAndMergeConfig("latest"), ",")
	if latest[1] != "ubuntu" {
		t.Fatalf("ReadAndMergeConfig(latest) = %q, want %q", latest[1], "ubuntu")
	}

	old := strings.Split(ReadAndMergeConfig("old"), ",")
	if old[1] != "debian" || old[2] != "/home/darbula/apps/monitor" {
		t.Fatalf("ReadAndMergeConfig(latest) = %q, want %q", old[1], "debian")
	}

	configFile = filepath.Join(config.GCODE_HOME, "gcode")
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if got, want := len(lines), 3; got != want {
		t.Fatalf("len(lines)=%d want=%d content=%q", got, want, string(content))
	}
}
