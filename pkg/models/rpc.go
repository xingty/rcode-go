package models

import "encoding/json"

var DELIMITER = byte(0x1e)

type SessionParams struct {
	Pid      int32  `json:"pid"`
	Hostname string `json:"hostname"`
	Keyfile  string `json:"keyfile"`
}

type OpenIDEParams struct {
	Sid  string `json:"sid"`
	Bin  string `json:"bin"`
	Path string `json:"path"`
}

type SessionPayload[T any] struct {
	Method string `json:"method"`
	Params T      `json:"params"`
}

type SessionData struct {
	Sid string `json:"sid"`
	Key string `json:"key"`
}

type MessageParams struct {
	Sid  string `json:"sid"`
	Skey string `json:"skey"`
	Bin  string `json:"bin"`
	Path string `json:"path"`
}

type MessagePayload struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type ResponsePayload[T any] struct {
	Code    int    `json:"code"`
	Data    T      `json:"data"`
	Message string `json:"message"`
}

func NewResponse[T any](code int, data T, message string) ResponsePayload[T] {
	return ResponsePayload[T]{
		Code:    code,
		Data:    data,
		Message: message,
	}
}

func NewRawResponse[T any](code int, data T, message string) []byte {
	res := NewResponse(code, data, message)
	jsondata, _ := json.Marshal(res)
	jsondata = append(jsondata, DELIMITER)

	return jsondata
}
