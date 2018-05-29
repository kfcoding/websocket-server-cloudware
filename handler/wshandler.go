package handler

import (
	"net/http"
	"log"
	"github.com/gorilla/websocket"
	"strings"
	"net/url"
	"sync"
)

var (
	tunnels  = make(map[string]*Tunnel)
	lock     sync.Mutex
	upgrader = websocket.Upgrader{}
)

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
		tunnel.Done <- tunnel.Id
		return
	}

	// get pulsar conn
	u := url.URL{Scheme: "ws", Host: tunnel.PodIP + ":9800", Path: "/"}
	pulsar, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		client.Close()
		log.Print("dial pulsar : ", err)
		tunnel.Done <- tunnel.Id
		return
	}

	// bind client and pulsar conn
	tunnel.Client = client
	tunnel.Pulsar = pulsar
	go tunnel.Iocopy()
	go tunnel.Iocopy2()
}

func addToTunnels(token string, tunnel *Tunnel) {
	lock.Lock()
	tunnels[token] = tunnel
	lock.Unlock()
}

func deleteFromTunnels(token string) {
	lock.Lock()
	delete(tunnels, token)
	lock.Unlock()
}

func Run(done chan string) {
	for {
		select {
		case token := <-done:
			if tunnel, ok := tunnels[token]; ok {
				deleteFromTunnels(token)
				log.Print(tunnel.Pod)
				//// call api to delete container
				//url := config.API_SERVER_ADDR + "/cloudware/deleteContainer"
				//req, _ := http.NewRequest("DELETE", url, nil)
				//req.Header.Add("podName", tunnel.Pod)
				//req.Header.Add("type", "0")
				//res, _ := http.DefaultClient.Do(req)
				//defer res.Body.Close()
				//body, _ := ioutil.ReadAll(res.Body)
				//fmt.Println(string(body))

				log.Print("delete tunnel succeed!")
			} else {
				log.Print("delete tunnel error!")
			}
		}
	}
}
