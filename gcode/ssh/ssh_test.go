package ssh

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/xingty/rcode-go/pkg/models"
)

func TestRemoteShellCommandStartsLoginShell(t *testing.T) {
	session := models.SessionData{Sid: "test-session", Key: "test-key"}
	want := `export RSSH_SID=test-session; export RSSH_SKEY=test-key; exec "$SHELL" -l`

	if got := remoteShellCommand(session); got != want {
		t.Fatalf("remoteShellCommand() = %q, want %q", got, want)
	}
}

func TestRemoteShellCommandLoadsLoginEnvironment(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("gssh's remote shell command targets Unix hosts")
	}

	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash is required to verify login shell startup")
	}

	home := t.TempDir()
	binDir := filepath.Join(home, "bin")
	if err := os.Mkdir(binDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Model a command installed in a PATH entry that is configured only for
	// login shells, as happens when batt is added from ~/.bash_profile.
	battPath := filepath.Join(binDir, "batt")
	if err := os.WriteFile(battPath, []byte("#!/bin/sh\nprintf 'batt-from-login-profile\\n'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	profile := "export PATH=\"$HOME/bin:$PATH\"\n"
	if err := os.WriteFile(filepath.Join(home, ".bash_profile"), []byte(profile), 0o600); err != nil {
		t.Fatal(err)
	}

	remoteCommand := remoteShellCommand(models.SessionData{
		Sid: "test-session",
		Key: "test-key",
	})
	cmd := exec.Command(bash, "--noprofile", "--norc", "-c", remoteCommand)
	cmd.Env = []string{
		"HOME=" + home,
		"PATH=/usr/bin:/bin",
		"SHELL=" + bash,
		"TERM=dumb",
	}
	cmd.Stdin = strings.NewReader("batt\nexit\n")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("remote shell failed: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), "batt-from-login-profile") {
		t.Fatalf("batt was not loaded from the login environment:\n%s", output)
	}
}
