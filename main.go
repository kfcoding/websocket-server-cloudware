package main

import (
	"flag"
	"log"
	"net/http"
	"github.com/websocket-server-cloudware/handler"
)

var addr = flag.String("addr", "0.0.0.0:8080", "http service address")

func main() {
	done := make(chan string, 1000)
	go handler.Run(done)

	handle, err := handler.CreateHTTPAPIHandler(done)
	if nil != err {
		log.Fatal(err)
	}

	flag.Parse()
	http.Handle("/api/", handle)
	http.HandleFunc("/api/websocket/cloudware/connect/", handler.Upgrade)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
