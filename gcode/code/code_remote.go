package code

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/djherbis/times"
	"github.com/xingty/rcode-go/gcode/config"
	"github.com/xingty/rcode-go/gcode/ipc"
	"github.com/xingty/rcode-go/pkg/models"
)

const MAX_IDLE_TIME = 4 * 60 * 60 * 100

var IS_RSSH_CLIENT = os.Getenv("RSSH_SID") != "" && os.Getenv("RSSH_SKEY") != ""

type FileInfo struct {
	Path  string
	Atime int64
}

func IsSocketOpen(addr string) bool {
	socks := ipc.NewIPCClientSocket(addr)
	return socks.Connect("unix") == nil
}

func GetCliPath(binName string) (string, error) {
	if !config.SUPPORTED_IDE.Has(binName) {
		return "", errors.New("unsupported ide")
	}

	binPath := binName
	if binName == "code" {
		binPath = "vscode"
	}

	homeDir, _ := os.UserHomeDir()
	codePath := fmt.Sprintf("%s/.%s-server/cli/servers", homeDir, binPath)
	servers, err := filepath.Glob(codePath + "/Stable-*")
	if err == nil && len(servers) > 0 {
		list := SortByAccessTime(servers)
		cli := list[0].Path + "/server/bin/remote-cli/" + binName
		return cli, nil
	}

	codePath = fmt.Sprintf("%s/.%s-server/bin", homeDir, binPath)
	servers, err = filepath.Glob(codePath + "/*")
	if err == nil && len(servers) > 0 {
		list := SortByAccessTime(servers)
		cli := list[0].Path + "/bin/remote-cli/" + binName
		return cli, nil
	}

	err = fmt.Errorf("can't find .%s-server at home dir. please install it fist", binName)
	return "", err
}

func IsRemote(binName string) (bool, error) {
	if !config.SUPPORTED_IDE.Has(binName) {
		return false, nil
	}

	return IS_RSSH_CLIENT || os.Getenv("SSH_CLIENT") != "", nil
}

func GetIpcSocket(binName string) (string, error) {
	if !config.SUPPORTED_IDE.Has(binName) {
		return "", errors.New("unsupported ide")
	}

	uid := os.Getuid()
	path := fmt.Sprintf("/run/user/%d/vscode-ipc-*.sock", uid)
	if runtime.GOOS == "darwin" {
		path = os.Getenv("TMPDIR") + "vscode-ipc-*.sock"
	}
	paths, err := filepath.Glob(path)
	if err != nil {
		return "", err
	}

	if len(paths) == 0 {
		return "", fmt.Errorf("can't find ipc socket")
	}

	return NextOpenSocket(SortByAccessTime(paths), binName)
}

func NextOpenSocket(list []FileInfo, binName string) (string, error) {
	now := time.Now().Unix()
	for _, info := range list {
		if now-info.Atime > MAX_IDLE_TIME {
			continue
		}

		if IsSocketOpen(info.Path) && IsSocketProcessRunning(info.Path, binName) {
			return info.Path, nil
		}
	}

	return "", os.ErrNotExist
}

func SortByAccessTime(paths []string) []FileInfo {
	list := make([]FileInfo, len(paths))
	for i, path := range paths {
		fp, _ := times.Stat(path)
		list[i] = FileInfo{
			Path:  path,
			Atime: int64(fp.AccessTime().Unix()),
		}
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Atime > list[j].Atime
	})

	return list
}

func IsSocketProcessRunning(sock string, binName string) bool {
	output, err := exec.Command("lsof", "-t", sock).Output()
	if err != nil {
		return false
	}

	pidStr := strings.TrimSpace(string(output))
	// data, err := os.ReadFile("/proc/" + pidStr + "/cmdline")
	data, err := exec.Command("ps", "-p", pidStr, "-o", "command").Output()
	if err != nil {
		return false
	}

	keyword := binName + "-server"
	return strings.Contains(string(data), keyword)
}

func sendMessage(binName string, dirName string, sid string, skey string) error {
	ipcSock := fmt.Sprintf("/tmp/rssh-ipc-%s.sock", sid)
	sock := ipc.NewIPCClientSocket(ipcSock)
	err := sock.Connect("unix")
	if err != nil {
		return err
	}

	defer sock.Close()
	params := models.OpenIDEParams{
		Sid:  sid,
		Skey: skey,
		Path: dirName,
		Bin:  binName,
	}

	rawParams, _ := json.Marshal(params)

	payload := models.MessagePayload{
		Method: "open_ide",
		Params: rawParams,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	err = sock.Send(data)
	if err != nil {
		return err
	}

	resData, err := sock.Receive()
	if err != nil {
		return err
	}

	res := &models.ResponsePayload[any]{}

	err = json.Unmarshal(resData, res)
	if err != nil {
		return err
	}

	if res.Code != 0 {
		return errors.New(res.Message)
	}

	return nil
}

func RunRemote(binName string, dirName string, maxIdleTime int) error {
	if len(dirName) == 0 {
		return fmt.Errorf(`need dir name here\n`)
	}

	stat, err := os.Stat(dirName)
	if err != nil {
		return err
	}

	openWithInnerIPC := func(binName string, dirName string) error {
		cli, err := GetCliPath(binName)
		if err != nil {
			return err
		}
		ipc_socket, err := GetIpcSocket(binName)
		if err != nil {
			return err
		}

		os.Setenv("VSCODE_IPC_HOOK_CLI", ipc_socket)
		err = exec.Command(cli, dirName, ipc_socket).Run()
		if err != nil {
			return err
		}

		return nil
	}

	openWithGSSHIPC := func(binName string, dirName string) error {
		// communicate with rssh's IPC Socket
		sid := os.Getenv("RSSH_SID")
		skey := os.Getenv("RSSH_SKEY")
		fmt.Println("running in gssh")

		err := sendMessage(binName, dirName, sid, skey)
		if err == nil {
			return nil
		}

		return err
	}

	if !stat.IsDir() && binName != "zed" {
		err := openWithInnerIPC(binName, dirName)
		if err != nil {
			return err
		}

		return nil
	}

	if !config.SUPPORTED_IDE.Has(binName) {
		return fmt.Errorf(`unsupported ide: %s\n`, binName)
	}

	if IS_RSSH_CLIENT {
		err := openWithGSSHIPC(binName, dirName)
		if err == nil {
			return nil
		}

		fmt.Printf("failed to send message: %s\ntrying fallback to vscode's IPC socket", err.Error())
	} else {
		fmt.Println("Warning: seems not running in gssh, trying fallback to vscode's IPC socket")
	}

	return openWithInnerIPC(binName, dirName)
}
