package ipc

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/google/uuid"
	"github.com/xingty/rcode-go/gcode/config"
	"github.com/xingty/rcode-go/pkg/models"
	"github.com/xingty/rcode-go/pkg/utils"
)

type Session struct {
	Pid      int32
	Hostname string
	addr     string
	Sid      string
	Skey     string
}

type MessageHandler struct {
	sessions map[string]*Session
	lock     sync.Mutex
}

func NewMessageHandler() *MessageHandler {
	return &MessageHandler{
		sessions: make(map[string]*Session),
	}
}

var rpc_methods = utils.NewSet("open_ide", "new_session")

func (h *MessageHandler) HandleMessage(rawData []byte) (any, error) {
	message := &models.MessagePayload{}
	err := json.Unmarshal(rawData, message)
	if err != nil {
		return nil, err
	}

	if !rpc_methods.Has(message.Method) {
		return nil, fmt.Errorf("unknown method: %s", message.Method)
	}

	switch message.Method {
	case "new_session":
		var sessionParams models.SessionParams
		err = json.Unmarshal(message.Params, &sessionParams)
		if err != nil {
			return nil, err
		}

		return h.NewSession(&sessionParams)

	case "open_ide":
		var ideParsms models.OpenIDEParams
		err = json.Unmarshal(message.Params, &ideParsms)
		if err != nil {
			return nil, err
		}

		return h.OpenIDE(&ideParsms)
	}

	return nil, fmt.Errorf("unknown method: %s", message.Method)
}

func doValidation(keyfile string, val string) error {
	key, err := os.ReadFile(keyfile)
	if err != nil {
		return err
	}

	if val != string(key) {
		return fmt.Errorf("invalid key")
	}

	return nil
}

func (h *MessageHandler) NewSession(params *models.SessionParams) (models.SessionData, error) {
	sid := uuid.New().String()
	skey := uuid.New().String()

	err := doValidation(config.GCODE_KEY_FILE, params.Keyfile)
	if err != nil {
		err := doValidation(config.RSSH_KEY_FILE, params.Keyfile)
		if err != nil {
			log.Printf("Authentication failed")
			return models.SessionData{}, err
		}
	}

	data := models.SessionData{
		Sid: sid,
		Key: skey,
	}

	h.lock.Lock()
	defer h.lock.Unlock()
	h.sessions[sid] = &Session{
		Pid:      params.Pid,
		Hostname: params.Hostname,
		Sid:      sid,
		Skey:     skey,
	}

	return data, nil
}

func (h *MessageHandler) OpenIDE(params *models.OpenIDEParams) (string, error) {
	if !config.SUPPORTED_IDE.Has(params.Bin) {
		return "", fmt.Errorf("unsupported ide")
	}

	if params.Sid == "" || params.Skey == "" {
		return "", fmt.Errorf("invalid session")
	}

	h.lock.Lock()
	session, ok := h.sessions[params.Sid]
	if !ok {
		h.lock.Unlock()
		return "", fmt.Errorf("invalid sid")
	}

	if params.Skey != session.Skey {
		h.lock.Unlock()
		return "", fmt.Errorf("invalid skey")
	}

	hostname := session.Hostname
	h.lock.Unlock()

	log.Printf("bin: %s, path: %s, hostname: %s\n", params.Bin, params.Path, hostname)

	binName := params.Bin
	path := params.Path
	uriType := "--folder-uri"
	if params.FileType == "file" {
		uriType = "--file-uri"
	}

	var ssh_remote string
	var cmd *exec.Cmd

	if binName == "zed" {
		ssh_remote = fmt.Sprintf("ssh://%s/%s", hostname, path)
		cmd = exec.Command(binName, ssh_remote)
	} else {
		ssh_remote = fmt.Sprintf("vscode-remote://ssh-remote+%s%s", hostname, path)
		cmd = exec.Command(binName, uriType, ssh_remote)
	}

	configureNoConsoleWindow(cmd)
	return "", cmd.Run()
}

func (h *MessageHandler) DestroySession(sid string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	delete(h.sessions, sid)
}

func (h *MessageHandler) SnapshotSessionPids() map[string]int32 {
	h.lock.Lock()
	defer h.lock.Unlock()

	out := make(map[string]int32, len(h.sessions))
	for sid, session := range h.sessions {
		out[sid] = session.Pid
	}
	return out
}
