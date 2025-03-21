package ssh

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/xingty/rcode-go/gcode/config"
	"github.com/xingty/rcode-go/gcode/ipc"
	"github.com/xingty/rcode-go/pkg/models"
)

func connect2IPCServer(ipc_host string, ipc_port int) *ipc.IPCClientSocket {
	addr := ipc_host + ":" + strconv.Itoa(ipc_port)
	sock := ipc.NewIPCClientSocket(addr)
	err := sock.Connect("tcp")
	if err == nil {
		return sock
	}

	fmt.Println("starting ipc server...")
	args := []string{"-host", ipc_host, "-port", strconv.Itoa(ipc_port)}
	err = ipc.StartIPCServer("gssh-ipc", args)
	if err != nil {
		panic(err)
	}
	time.Sleep(100 * time.Millisecond)

	for i := 1; i < 10; i++ {
		sock = ipc.NewIPCClientSocket(addr)
		err := sock.Connect("tcp")
		if err == nil {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	return sock
}

func createSession(sock *ipc.IPCClientSocket, hostname string) models.SessionData {
	data, err := os.ReadFile(config.RSSH_KEY_FILE)
	if err != nil {
		data, err = os.ReadFile(config.GCODE_KEY_FILE)
		if err != nil {
			panic(err)
		}
	}

	session := models.SessionPayload[models.SessionParams]{
		Method: "new_session",
		Params: models.SessionParams{
			Pid:      int32(os.Getpid()),
			Hostname: hostname,
			Keyfile:  string(data),
		},
	}

	jsondata, err := json.Marshal(session)
	if err != nil {
		panic(err)
	}

	err = sock.Send(jsondata)
	if err != nil {
		panic(err)
	}

	response, err := sock.Receive()
	if err != nil {
		panic(err)
	}

	res := models.ResponsePayload[models.SessionData]{}
	json.Unmarshal(response, &res)

	return res.Data
}

func findHostPos(args []string) int {
	for i, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			return i
		}
	}

	return -1
}

func createSSHArgs(
	host string,
	port int,
	ssh_args []string) []string {

	pseudo := false
	for i := range ssh_args {
		param := ssh_args[i]
		if param == "-R" || param == "-T" {
			fmt.Println("Warning: gssh is disabled because of -R or -T")
			fmt.Println("ssh is used instead")
			return ssh_args
		}

		if param == "-t" {
			pseudo = true
		}
	}

	index := findHostPos(ssh_args)
	if index == -1 {
		fmt.Println("Error: host not found")
		os.Exit(1)
	}

	pre := ssh_args[:index]
	post := ssh_args[index:]
	hostname := ssh_args[index]

	socks := connect2IPCServer(host, port)
	s := createSession(socks, hostname)
	socks.Close()

	buf := make([]string, 0)
	buf = append(buf, pre...)
	if !pseudo {
		buf = append(buf, "-t")
	}

	sock := fmt.Sprintf("/tmp/rssh-ipc-%s.sock", s.Sid)
	tunnel := fmt.Sprintf("%s:%s:%d", sock, host, port)
	buf = append(buf, "-R", tunnel)
	buf = append(buf, post...)
	env := fmt.Sprintf("export RSSH_SID=%s; export RSSH_SKEY=%s; exec $SHELL", s.Sid, s.Key)

	return append(buf, env)
}

func Run(ipc_host string, ipc_port int, ssh_args []string) {
	newArgs := createSSHArgs(ipc_host, ipc_port, ssh_args)
	ipc.StartSSHClient(newArgs)
}
