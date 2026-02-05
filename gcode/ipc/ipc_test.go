package ipc

import (
	"testing"

	"github.com/xingty/rcode-go/pkg/models"
)

func TestOpenIDERejectsInvalidSkey(t *testing.T) {
	h := NewMessageHandler()

	h.lock.Lock()
	h.sessions["sid1"] = &Session{
		Pid:      123,
		Hostname: "example",
		Sid:      "sid1",
		Skey:     "good",
	}
	h.lock.Unlock()

	_, err := h.OpenIDE(&models.OpenIDEParams{
		Sid:      "sid1",
		Skey:     "bad",
		Bin:      "code",
		Path:     "/tmp",
		FileType: "dir",
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != "invalid skey" {
		t.Fatalf("unexpected error: %v", err)
	}
}
