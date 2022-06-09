package tcp

import (
	"net"
	"sync"
)

type EchoHandler struct {
	activeConn sync.Map
	closing    atomic.Boolen
}

func MakeEchoHandler() *EchoHandler {
	return &EchoHandler{}
}

type EchoClient struct {
	Conn    net.Conn
	Waiting wait.Wait
}
