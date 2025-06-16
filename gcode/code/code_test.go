package code

import (
	"os"
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
	isRemote, _ := IsRemote("code")

	_, err := GetCliPath("code")
	if err != nil && isRemote {
		t.Error(err)
	}
}

func TestExpandDir(t *testing.T) {
	home, _ := os.UserHomeDir()
	expanded, err := expandDir("localhost", "~/test", "linux")
	if err != nil {
		t.Error(err)
	}

	expected := home + "/test"
	if expanded != expected {
		t.Errorf("expanded: %s, expected: %s", expanded, expected)
	}
}
