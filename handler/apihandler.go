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
)

type APIHandler struct {
	Done chan string
}

type CloudwareResponse struct {
	Id string `json:"id"`
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
		apiV1Ws.GET("/websocket/cloudware/gettoken/{namespace}/{pod}/{container}").
			To(apiHandler.handleCloudware).
			Writes(CloudwareResponse{}))

	return wsContainer, nil
}

func (apiHandler *APIHandler) handleCloudware(request *restful.Request, response *restful.Response) {
	// gen token
	sessionId, err := genTerminalSessionId()
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	// add to tunnels
	timer := make(chan bool)
	addToTunnels(sessionId, &Tunnel{
		Id:    sessionId,
		Done:  apiHandler.Done,
		Timer: timer,
	})

	// return result
	response.Header().Set("Access-Control-Allow-Origin", "*")
	response.WriteHeaderAndEntity(http.StatusOK, CloudwareResponse{Id: sessionId})

	go func() {
		// add a timer, if client not connect ws in a minute, channel will be delete
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)
		ticker := time.NewTicker(time.Second * 10)
		defer func() {
			ticker.Stop()
			close(timer)
			log.Print("terminal thread")
		}()
		select {
		case <-timer:
			return
		case <-ticker.C:
			log.Print("time out: ", sessionId)
			deleteFromTunnels(sessionId)
			return
		case <-interrupt:
			log.Println("system interrupt")
			return
		}
	}()
}

func genTerminalSessionId() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	id := make([]byte, hex.EncodedLen(len(bytes)))
	hex.Encode(id, bytes)
	return string(id), nil
}
