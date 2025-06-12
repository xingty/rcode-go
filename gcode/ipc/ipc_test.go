package ipc

import "testing"

type MockMessageHandler struct {
	sessions map[string]*Session
}

func (m *MockMessageHandler) HandleMessage(rawData []byte) (any, error) {
	return nil, nil
}

func TestIPCClientSocket(t *testing.T) {}
