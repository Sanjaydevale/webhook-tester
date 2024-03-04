package cli

import (
	"fmt"
	"io"
	"log"
	"time"
	"whtester/serialize"

	"github.com/gorilla/websocket"
)

type client struct {
	URL  string
	Conn *websocket.Conn
}

func (c *client) PrintMessage(w io.Writer) {

	data, err := read(c.Conn)
	if err != nil {
		return
	}
	fmt.Fprint(w, data)

}

func Newclient() *client {
	c := &client{}
	c.Conn = NewConn("ws://localhost:8080/ws")
	c.URL = readURL(c.Conn)
	return c
}

func read(ws *websocket.Conn) (string, error) {
	for {
		msgType, data, err := ws.ReadMessage()
		if err != nil {
			return "", err
		}
		if len(data) != 0 {
			if msgType == websocket.TextMessage {
				return "\n" + string(data), nil
			} else if msgType == websocket.BinaryMessage {
				req := serialize.DecodeRequest(data)
				body, err := io.ReadAll(req.Body)
				if err != nil {
					log.Fatalf("error reading body of request, got error %v", err)
				}
				method := req.Method
				return fmt.Sprintf("Body :%s\nMethod :%s", string(body), method), nil
			}
		}
	}
}

func readURL(ws *websocket.Conn) string {
	result := make(chan string, 1)
	go func() {
		data, _ := read(ws)
		result <- data
		close(result)
	}()
	select {
	case url := <-result:
		return url
	case <-time.After(5 * time.Second):
		log.Fatalf("took too long to read message from server")
	}
	return ""
}

func NewConn(wsLink string) *websocket.Conn {
	ws, _, err := websocket.DefaultDialer.Dial(wsLink, nil)
	if err != nil {
		log.Fatalf("error establishing websocket connection: %v", err.Error())
	}
	return ws
}
