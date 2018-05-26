package handler

import (
	"net/http"
	"log"
	"github.com/gorilla/websocket"
	"strings"
	"net/url"
)

var upgrader = websocket.Upgrader{} // use default options

func Upgrade(w http.ResponseWriter, r *http.Request) {
	var (
		client *websocket.Conn
		tunnel *Tunnel
		ok     bool
		err    error
	)

	// get tunnel
	paths := strings.Split(r.URL.Path, "/")
	token := paths[len(paths)-1]
	if tunnel, ok = tunnels[token]; !ok {
		log.Printf("handleSession: can't find session '%s'", token)
		return
	}
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	tunnel.Timer <- true
	if tunnel, ok = tunnels[token]; !ok {
		log.Printf("handleSession: can't find session '%s'", token)
		return
	}

	// get client conn
	client, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	// get pulsar conn
	u := url.URL{Scheme: "ws", Host: "localhost:8081", Path: "/pulsar"}
	pulsar, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		client.Close()
		log.Fatal("dial pulsar : ", err)
	}

	tunnel.Client = client
	tunnel.Pulsar = pulsar
	go tunnel.Iocopy()
	go tunnel.Iocopy2()
}
