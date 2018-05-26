package handler

import (
	"net/http"
	"github.com/emicklei/go-restful"
	"encoding/hex"
	"crypto/rand"
	"os"
	"os/signal"
	"time"
	"log"
	"github.com/websocket-server-cloudware/config"
)

type APIHandler struct {
	Done chan string
}

func CreateHTTPAPIHandler(done chan string) (http.Handler, error) {
	apiHandler := APIHandler{Done: done}

	wsContainer := restful.NewContainer()
	wsContainer.EnableContentEncoding(true)

	apiV1Ws := new(restful.WebService)

	apiV1Ws.Path("/api").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	wsContainer.Add(apiV1Ws)

	apiV1Ws.Route(
		apiV1Ws.GET("/websocket/getws/{pod}/{ip}").
			To(apiHandler.handleCloudware))

	return wsContainer, nil
}

func (apiHandler *APIHandler) handleCloudware(request *restful.Request, response *restful.Response) {
	// get params from request path
	var (
		pod = request.PathParameter("pod")
		ip  = request.PathParameter("ip")
	)

	// gen token
	token, err := genCloudwareSessionId()
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	// add to tunnels
	timer := make(chan bool)
	addToTunnels(pod, token, &Tunnel{
		Id:    token,
		Done:  apiHandler.Done,
		Timer: timer,
		Pod:   pod,
		PodIP: ip,
	})

	// return result
	response.Header().Set("Access-Control-Allow-Origin", "*")
	response.WriteHeaderAndEntity(http.StatusOK, "ws://"+config.CLOUDWARE_WSS_DOMAIN+"/api/websocket/connect/"+token)

	// add a timer, if client not connect ws in a minute, channel will be delete
	go func() {
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)
		ticker := time.NewTicker(time.Second * 10)
		defer func() {
			ticker.Stop()
			close(timer)
			log.Print("terminal thread : ", token)
		}()
		select {
		case <-timer:
			return
		case <-ticker.C:
			log.Print("time out: ", token)
			deleteFromTunnels(pod, token)
			return
		case <-interrupt:
			log.Println("system interrupt")
			return
		}
	}()
}

func genCloudwareSessionId() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	id := make([]byte, hex.EncodedLen(len(bytes)))
	hex.Encode(id, bytes)
	return string(id), nil
}
