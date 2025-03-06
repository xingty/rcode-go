package code

import (
	"os"
)

const MAX_IDLE_TIME = 4 * 60 * 60

var IS_RSSH_CLIENT = os.Getenv("RSSH_SID") != "" && os.Getenv("RSSH_SKEY") != ""

func runLocal() {

}

func runRemote(binName string, dirName string, maxIdleTime int) {

	println("starting ipc server...")
	println("starting ssh client...")
}

func StartCode(binName string, remoteName string) {

}
