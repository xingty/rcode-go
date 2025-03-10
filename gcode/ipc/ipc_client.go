package ipc

import (
	"bytes"
	"errors"
	"net"

	"github.com/xingty/rcode-go/pkg/models"
)

type IPCClientSocket struct {
	addr string
	conn net.Conn
}

func NewIPCClientSocket(addr string) *IPCClientSocket {
	return &IPCClientSocket{addr: addr, conn: nil}
}

func (s *IPCClientSocket) Connect(network string) error {
	if s.conn != nil {
		return errors.New("already connected")
	}

	conn, err := net.Dial(network, s.addr)
	if err != nil {
		return errors.New("failed to connect to RPC server")
	}

	s.conn = conn
	return nil
}

func (s *IPCClientSocket) Send(data []byte) error {
	if s.conn == nil {
		return errors.New("not connected")
	}

	if data[len(data)-1] != models.DELIMITER {
		data = append(data, models.DELIMITER)
	}

	_, err := s.conn.Write(data)
	return err
}

func (s *IPCClientSocket) Receive() ([]byte, error) {
	if s.conn == nil {
		return nil, errors.New("not connected")
	}

	buf := make([]byte, 0)
	data := make([]byte, 1024)
	delimiter := []byte{models.DELIMITER}

	for {
		n, err := s.conn.Read(data)
		if err != nil {
			return nil, err
		}

		if n == 0 {
			return nil, errors.New("no data received")
		}

		index := bytes.Index(data, delimiter)
		if index != -1 {
			buf = append(buf, data[:index]...)
			break
		} else {
			buf = append(buf, data[:n]...)
		}
	}

	return buf, nil
}

func (s *IPCClientSocket) Close() error {
	if s.conn == nil {
		return errors.New("no connection to close")
	}

	return s.conn.Close()
}

func (s *IPCClientSocket) Read(b []byte) (int, error) {
	if s.conn == nil {
		return 0, errors.New("not connected")
	}

	return s.conn.Read(b)
}
