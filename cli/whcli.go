package cli

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"time"
	"whtester/serialize"

	"github.com/gorilla/websocket"
)

type client struct {
	URL  string
	Conn *websocket.Conn
}

func (c *client) PrintMessage(w io.Writer, fields []string) error {

	data, err := read(c.Conn, fields)
	if err != nil {
		return err
	}
	fmt.Fprint(w, data)
	return nil
}

func Newclient() *client {
	c := &client{}
	c.Conn = NewConn("ws://localhost:8080/ws")
	c.URL = readURL(c.Conn)
	return c
}

func read(ws *websocket.Conn, fields []string) (string, error) {
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
				res := readRequestFields(fields, *req)
				return res, nil
			}
		}
	}
}

func readRequestFields(fields []string, req http.Request) string {
	out := ""
	r := reflect.ValueOf(req)
	for _, f := range fields {
		if r.FieldByName(f) == reflect.ValueOf(nil) {
			fmt.Printf("does not have field, %s", f)
		}
		field := fmt.Sprintf("\n%s :%v", f, r.FieldByName(f).Interface())
		out += field
	}
	return out
}

func readURL(ws *websocket.Conn) string {
	result := make(chan string, 1)
	go func() {
		data, _ := read(ws, nil)
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
