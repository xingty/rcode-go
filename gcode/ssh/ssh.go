package ssh

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/xingty/rcode-go/gcode/config"
	"github.com/xingty/rcode-go/gcode/ipc"
	"github.com/xingty/rcode-go/pkg/models"
)

func connect2IPCServer(ipc_host string, ipc_port int) *ipc.IPCClientSocket {
	sock := ipc.NewIPCClientSocket(ipc_host, ipc_port)
	err := sock.Connect("tcp")
	if err == nil {
		return sock
	}

	fmt.Println("starting ipc server...")
	args := []string{"--host", ipc_host, "--port", strconv.Itoa(ipc_port)}
	ipc.StartIPCServer("gssh-ipc", args)
	time.Sleep(100 * time.Millisecond)
	sock.Close()

	for i := 1; i < 10; i++ {
		sock = ipc.NewIPCClientSocket(ipc_host, ipc_port)
		err := sock.Connect("tcp")
		if err == nil {
			break
		}

		if !errors.Is(err, syscall.ECONNREFUSED) {
			panic(err)
		}

		sock = ipc.NewIPCClientSocket(ipc_host, ipc_port)
		time.Sleep(100 * time.Millisecond)
	}

	return sock
}

func createSession(sock *ipc.IPCClientSocket, hostname string) models.SessionData {
	defer sock.Close()

	data, err := os.ReadFile(config.GCODE_KEY_FILE)
	if err != nil {
		panic(err)
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
			fmt.Printf("Error: %s is not allowed\n", param)
			os.Exit(1)
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

	buf := make([]string, 0)
	buf = append(buf, pre...)
	if !pseudo {
		buf = append(buf, "-t")
	}

	sock := fmt.Sprintf("/tmp/gssh-ipc-%s.sock", s.Sid)
	tunnel := fmt.Sprintf("%s:%s:%d", sock, host, port)
	buf = append(buf, "-R", tunnel)
	buf = append(buf, post...)
	env := fmt.Sprintf("export RSSH_SID=%s; export RSSH_SKEY=%s; exec $SHELL", s.Sid, s.Key)

	return append(buf, env)
}

func Run(ipc_host string, ipc_port int, ssh_args []string) {
	println("starting ipc server...")
	println("starting ssh client...")

	newArgs := createSSHArgs(ipc_host, ipc_port, ssh_args)
	fmt.Printf("ssh args: %v\n", newArgs)
	ipc.StartSSHClient(newArgs)
}
