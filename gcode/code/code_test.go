package code

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsRemote(t *testing.T) {
	_, err := IsRemote("code")
	if err != nil {
		t.Error(err)
	}

	os.Setenv("SSH_CLIENT", "127.0.0.1")
	isRemote, err := IsRemote("code")
	if err != nil {
		t.Error(err)
	}

	if !isRemote {
		t.Errorf("isRemote: %t, except: true", isRemote)
	}

	t.Logf("remote: %v", isRemote)
}

func TestGetCliPath(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("SSH_CLIENT", "")
	t.Setenv("RSSH_SID", "")
	t.Setenv("RSSH_SKEY", "")

	cli := filepath.Join(
		tempHome,
		".vscode-server",
		"cli",
		"servers",
		"Stable-1",
		"server",
		"bin",
		"remote-cli",
		"code",
	)
	if err := os.MkdirAll(filepath.Dir(cli), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(cli, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write: %v", err)
	}

	got, err := GetCliPath("code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != cli {
		t.Fatalf("got=%q want=%q", got, cli)
	}
}

func TestExpandDir(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	sshCfg := filepath.Join(tempHome, ".ssh", "config")
	if err := os.MkdirAll(filepath.Dir(sshCfg), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(sshCfg, []byte("Host localhost\n  User demo\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	expanded, err := expandDir("localhost", "~/test", "linux")
	if err != nil {
		t.Error(err)
	}

	expected := "/home/demo/test"
	if expanded != expected {
		t.Errorf("expanded: %s, expected: %s", expanded, expected)
	}
}
