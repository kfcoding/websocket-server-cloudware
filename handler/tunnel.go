package handler

import (
	"github.com/gorilla/websocket"
	"sync"
	"log"
)

type Tunnel struct {
	Id     string
	Pulsar *websocket.Conn
	Client *websocket.Conn
	Done   chan string
	Timer  chan bool
}

var lock sync.Mutex
var tunnels = make(map[string]*Tunnel)
var cloudwares = make(map[string]int)

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

func addToTunnels(key string, tunnel *Tunnel) {
	lock.Lock()
	tunnels[key] = tunnel
	lock.Unlock()
}

func deleteFromTunnels(key string) {
	lock.Lock()
	delete(tunnels, key)
	lock.Unlock()
}

func Run(done chan string) {
	for {
		select {
		case token := <-done:
			log.Print(token) //	websocket disconnected
		}
	}
}
