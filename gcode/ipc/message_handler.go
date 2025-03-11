package ipc

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"

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
}

type MessageHandler struct {
	sessions map[string]*Session
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
		log.Printf("Authentication failed, key: %s", params.Keyfile)
		if err != nil {
			return models.SessionData{}, err
		}
	}

	data := models.SessionData{
		Sid: sid,
		Key: skey,
	}

	h.sessions[sid] = &Session{
		Pid:      params.Pid,
		Hostname: params.Hostname,
		Sid:      sid,
	}

	return data, nil
}

func (h *MessageHandler) OpenIDE(params *models.OpenIDEParams) (string, error) {
	if !config.SUPPORTED_IDE.Has(params.Bin) {
		return "", fmt.Errorf("unsupported ide")
	}

	session, ok := h.sessions[params.Sid]
	if !ok {
		return "", fmt.Errorf("invalid sid")
	}

	log.Printf("bin: %s, path: %s, hostname: %s\n", params.Bin, params.Path, session.Hostname)

	binName := params.Bin
	hostname := session.Hostname
	path := params.Path

	ssh_remote := fmt.Sprintf("vscode-remote://ssh-remote+%s%s", hostname, path)
	cmd := exec.Command(binName, "--folder-uri", ssh_remote)

	err := cmd.Start()
	if err != nil {
		return "", err
	}

	cmd.Run()
	return "", nil
}

func (h *MessageHandler) DestroySession(sid string) {
	delete(h.sessions, sid)
}
