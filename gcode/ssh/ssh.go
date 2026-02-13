package ssh

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/xingty/rcode-go/gcode/config"
	"github.com/xingty/rcode-go/gcode/ipc"
	"github.com/xingty/rcode-go/pkg/models"
)

const (
	connectRetryDelay = 100 * time.Millisecond
	connectMaxRetries = 10
)

var errHostNotFound = errors.New("host not found")

func connectToIPCServer(ipcHost string, ipcPort int) (*ipc.IPCClientSocket, error) {
	addr := ipcHost + ":" + strconv.Itoa(ipcPort)

	sock := ipc.NewIPCClientSocket(addr)
	if err := sock.Connect("tcp"); err == nil {
		return sock, nil
	}

	fmt.Fprintln(os.Stderr, "starting ipc server...")
	args := []string{"-host", ipcHost, "-port", strconv.Itoa(ipcPort)}
	if err := ipc.StartIPCServer("gssh-ipc", args); err != nil {
		return nil, err
	}

	var lastErr error
	for i := 0; i < connectMaxRetries; i++ {
		time.Sleep(connectRetryDelay)
		sock = ipc.NewIPCClientSocket(addr)
		if err := sock.Connect("tcp"); err == nil {
			return sock, nil
		} else {
			lastErr = err
		}
	}

	if lastErr == nil {
		lastErr = errors.New("unknown connection failure")
	}
	return nil, fmt.Errorf("failed to connect to ipc server at %s: %w", addr, lastErr)
}

func createSession(sock *ipc.IPCClientSocket, hostname string, ownerPID int32) (models.SessionData, error) {
	if ownerPID <= 0 {
		ownerPID = int32(os.Getpid())
	}

	data, err := os.ReadFile(config.RSSH_KEY_FILE)
	if err != nil {
		data, err = os.ReadFile(config.GCODE_KEY_FILE)
		if err != nil {
			return models.SessionData{}, err
		}
	}

	session := models.SessionPayload[models.SessionParams]{
		Method: "new_session",
		Params: models.SessionParams{
			Pid:      ownerPID,
			Hostname: hostname,
			Keyfile:  string(data),
		},
	}

	jsondata, err := json.Marshal(session)
	if err != nil {
		return models.SessionData{}, err
	}

	if err := sock.Send(jsondata); err != nil {
		return models.SessionData{}, err
	}

	response, err := sock.Receive()
	if err != nil {
		return models.SessionData{}, err
	}

	res := models.ResponsePayload[models.SessionData]{}
	if err := json.Unmarshal(response, &res); err != nil {
		return models.SessionData{}, err
	}
	if res.Code != 0 {
		if res.Message == "" {
			res.Message = "unknown ipc error"
		}
		return models.SessionData{}, errors.New(res.Message)
	}

	return res.Data, nil
}

type sshInvocation struct {
	destinationIndex int
	hasPseudoTTY     bool
	disableReason    string // non-empty means "fallback to plain ssh"
}

func analyzeSSHArgs(args []string) (sshInvocation, error) {
	inv := sshInvocation{destinationIndex: -1}

	endOfOptions := false
	skipNext := false

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if skipNext {
			skipNext = false
			continue
		}

		if endOfOptions || arg == "" || arg == "-" || !strings.HasPrefix(arg, "-") {
			inv.destinationIndex = i
			break
		}

		if arg == "--" {
			endOfOptions = true
			continue
		}

		// ssh doesn't really use long options; treat unknown `--foo` as an option with no value.
		if strings.HasPrefix(arg, "--") {
			continue
		}

		hasTTY, forbidden, needsValue := scanShortOptionToken(arg)
		if hasTTY {
			inv.hasPseudoTTY = true
		}
		if forbidden != "" && inv.disableReason == "" {
			inv.disableReason = forbidden
		}
		if needsValue {
			skipNext = true
		}
	}

	if inv.destinationIndex == -1 {
		if inv.disableReason != "" {
			return inv, nil
		}
		return inv, errHostNotFound
	}

	if inv.disableReason == "" && inv.destinationIndex+1 < len(args) {
		inv.disableReason = "remote command"
	}

	return inv, nil
}

func scanShortOptionToken(token string) (hasTTY bool, forbidden string, needsValue bool) {
	if !strings.HasPrefix(token, "-") || token == "-" || token == "--" {
		return false, "", false
	}

	opts := token[1:]
	for i := 0; i < len(opts); i++ {
		c := opts[i]
		switch c {
		case 'G':
			// `ssh -G host` prints configuration and exits without connecting.
			// Avoid creating a session for it.
			return hasTTY, "-G", false
		case 'V':
			// `ssh -V` prints version and exits.
			return hasTTY, "-V", false
		case 'Q':
			// `ssh -Q` queries algorithms and exits.
			return hasTTY, "-Q", true
		case 'h', '?':
			// `ssh -h` / `ssh -?` prints usage and exits.
			return hasTTY, "-h", false
		case 't':
			hasTTY = true
		case 'T':
			return hasTTY, "-T", false
		case 'R':
			return hasTTY, "-R", false
		default:
			if !shortOptionRequiresValue(c) {
				continue
			}

			// Option takes a value: if attached (`-p22`, `-oFoo=bar`, `-i/path`) we don't need to
			// consume the next arg; otherwise we do.
			if i < len(opts)-1 {
				return hasTTY, "", false
			}
			return hasTTY, "", true
		}
	}

	return hasTTY, "", false
}

func shortOptionRequiresValue(opt byte) bool {
	switch opt {
	case 'b', 'c', 'D', 'E', 'e', 'F', 'I', 'i', 'J', 'L', 'l', 'm', 'O', 'o', 'p', 'Q', 'S', 'W', 'w':
		return true
	default:
		return false
	}
}

func BuildSSHArgs(ipcHost string, ipcPort int, ownerPID int32, sshArgs []string) ([]string, error) {
	inv, err := analyzeSSHArgs(sshArgs)
	if err != nil {
		if errors.Is(err, errHostNotFound) {
			// Some ssh invocations (e.g. `ssh -h`, `ssh -V`) are meaningful without a hostname.
			// In those cases we should not create a session; just run plain ssh.
			return sshArgs, nil
		}
		return nil, err
	}

	if inv.disableReason != "" {
		switch inv.disableReason {
		case "-R", "-T":
			fmt.Fprintf(os.Stderr, "Warning: gssh is disabled because of %s; using ssh instead\n", inv.disableReason)
		case "remote command":
			fmt.Fprintln(os.Stderr, "Warning: gssh is disabled because remote command is provided; using ssh instead")
		default:
			fmt.Fprintf(os.Stderr, "Warning: gssh is disabled (%s); using ssh instead\n", inv.disableReason)
		}
		return sshArgs, nil
	}

	pre := sshArgs[:inv.destinationIndex]
	post := sshArgs[inv.destinationIndex:]
	hostname := sshArgs[inv.destinationIndex]

	sock, err := connectToIPCServer(ipcHost, ipcPort)
	if err != nil {
		return nil, err
	}
	defer func() { _ = sock.Close() }()

	s, err := createSession(sock, hostname, ownerPID)
	if err != nil {
		return nil, err
	}

	out := make([]string, 0, len(sshArgs)+6)
	out = append(out, pre...)
	if !inv.hasPseudoTTY {
		out = append(out, "-t")
	}

	remoteSock := fmt.Sprintf("/tmp/rssh-ipc-%s.sock", s.Sid)
	tunnel := fmt.Sprintf("%s:%s:%d", remoteSock, ipcHost, ipcPort)
	out = append(out, "-R", tunnel)
	out = append(out, post...)

	env := fmt.Sprintf("export RSSH_SID=%s; export RSSH_SKEY=%s; exec $SHELL", s.Sid, s.Key)
	out = append(out, env)

	return out, nil
}

func Run(ipcHost string, ipcPort int, sshArgs []string) int {
	newArgs, err := BuildSSHArgs(ipcHost, ipcPort, int32(os.Getpid()), sshArgs)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 255
	}
	return ipc.StartSSHClient(newArgs)
}
