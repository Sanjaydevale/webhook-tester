package cli

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type client struct {
	URL  string
	Conn *websocket.Conn
}

func (c *client) PrintMessage(w io.Writer) {
	for {
		fmt.Fprintf(w, "%v", read(c.Conn))
	}
}

func Newclient() *client {
	c := &client{}
	c.Conn = NewConn()
	c.URL = readURL(c.Conn)
	return c
}

func read(ws *websocket.Conn) string {
	for {
		_, data, err := ws.ReadMessage()
		if err != nil {
			continue
		}
		if len(data) != 0 {
			return string(data)
		}
	}
}

func readURL(ws *websocket.Conn) string {
	result := make(chan string, 1)
	select {
	case result <- read(ws):
		return <-result
	case <-time.After(5 * time.Second):
		log.Fatalf("took too long to read message from server")
	}
	return ""
}

func NewConn() *websocket.Conn {
	ws, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		log.Fatalf("error establishing websocket connection: %v", err.Error())
	}
	return ws
}
