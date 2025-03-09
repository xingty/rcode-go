package ipc

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/samber/lo"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/xingty/rcode-go/pkg/models"
	"github.com/xingty/rcode-go/pkg/utils"
)

type IPCServerSocket struct {
	handler     *MessageHandler
	maxIdleTime int
	idle        int
	done        chan struct{}
}

func NewIPCServerSocket(maxIdleTime int) *IPCServerSocket {
	return &IPCServerSocket{
		handler:     NewMessageHandler(),
		maxIdleTime: maxIdleTime,
		idle:        0,
		done:        make(chan struct{}),
	}
}

func (s *IPCServerSocket) handleClient(conn net.Conn) error {
	data := make([]byte, 1024)
	delimiter := []byte{models.DELIMITER}
	buf := make([]byte, 0)

	for {
		n, err := conn.Read(data)
		if err != nil {
			return err
		}

		if n == 0 {
			return errors.New("no data received")
		}

		index := bytes.Index(data, delimiter)
		if index != -1 {
			buf = append(buf, data[:index]...)
			data, err := s.handler.HandleMessage(buf)
			if err != nil {
				fmt.Printf("%s", err.Error())
				rawData := models.NewRawResponse(1, "", err.Error())
				conn.Write(rawData)
				return err
			}

			resData := models.NewRawResponse(0, data, "")
			conn.Write(resData)

			return nil
		} else {
			buf = append(buf, data[:n]...)
		}
	}
}

func (s *IPCServerSocket) handleConnection(listener net.Listener) {
	defer close(s.done)

	idle := 0

	for {
		tcpListener, ok := listener.(*net.TCPListener)
		if !ok {
			log.Fatal("无法将监听器转换为TCPListener")
		}

		tcpListener.SetDeadline(time.Now().Add(10 * time.Second))
		conn, err := tcpListener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				log.Println("监听器已关闭，退出handleConnection")
				return
			}

			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				activeSessions, inactiveSessions := s.getSessions()
				clients := len(activeSessions) + len(inactiveSessions)
				fmt.Printf("active: %d, inactive: %d, idle=%d\n", len(activeSessions), len(inactiveSessions), idle)
				if len(inactiveSessions) > 0 {
					for _, sid := range inactiveSessions {
						s.handler.DestroySession(sid)
					}
				}

				if clients > 0 {
					idle = 0
				} else {
					idle += 10
				}

				if clients == 0 && idle > s.maxIdleTime {
					fmt.Println("Server stopped: idle", idle)
					return
				}

				continue
			}
			log.Printf("接受连接时出错: %v", err)
			continue
		}

		go s.handleClient(conn)
	}
}

func (s *IPCServerSocket) getSessions() ([]string, []string) {
	curSessions := s.handler.sessions
	activeSessions := make([]string, 0)
	inactiveSessions := make([]string, 0)
	if len(curSessions) == 0 {
		return activeSessions, inactiveSessions
	}

	pids, err := process.Pids()
	if err != nil {
		return lo.Keys(curSessions), inactiveSessions
	}

	pidSet := utils.NewSet(pids...)
	for sid, session := range curSessions {
		if pidSet.Has(session.Pid) {
			activeSessions = append(activeSessions, sid)
		} else {
			inactiveSessions = append(inactiveSessions, sid)
		}
	}

	return activeSessions, inactiveSessions
}

func (s *IPCServerSocket) Start(host string, port int) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	addr := fmt.Sprintf("%s:%d", host, port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer listener.Close()
	fmt.Println("Server listening on ", listener.Addr())

	go s.handleConnection(listener)

	select {
	case <-sigChan:
		return nil
	case <-s.done:
		return nil
	}
}

func (s *IPCServerSocket) Stop() error {
	defer close(s.done)
	return nil
}
