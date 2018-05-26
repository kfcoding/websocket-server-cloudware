package handler

import (
	"github.com/gorilla/websocket"
)

type Tunnel struct {
	Id     string
	Pulsar *websocket.Conn
	Client *websocket.Conn
	Done   chan string
	Timer  chan bool
	Pod    string
	PodIP  string
}

// from client to pulsar
func (tunnel *Tunnel) Iocopy() {
	defer func() {
		tunnel.Pulsar.Close()
		tunnel.Client.Close()
		tunnel.Done <- tunnel.Id
	}()
	for {
		_, msg, e := tunnel.Client.ReadMessage()
		if nil != e {
			return
		}
		tunnel.Pulsar.WriteMessage(websocket.TextMessage, msg)
	}
}

// from pulsar to client
func (tunnel *Tunnel) Iocopy2() {
	defer func() {
		tunnel.Pulsar.Close()
		tunnel.Client.Close()
	}()
	for {
		_, msg, e := tunnel.Pulsar.ReadMessage()
		if nil != e {
			return
		}
		tunnel.Client.WriteMessage(websocket.TextMessage, msg)
	}
}
