package main

import (
	"flag"
	"log"
	"net/http"
	"github.com/websocket-server-cloudware/handler"
	"github.com/websocket-server-cloudware/config"
)

var addr = flag.String("addr", config.SERVER_ADDRESS, "http service address")

func main() {
	done := make(chan string, 1000)
	go handler.Run(done)

	handle, err := handler.CreateHTTPAPIHandler(done)
	if nil != err {
		log.Fatal(err)
	}

	flag.Parse()
	http.Handle("/api/", handle)
	http.HandleFunc("/api/websocket/connect/", handler.Upgrade)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
