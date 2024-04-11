package cli

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"time"
	"whtester/serialize"

	"github.com/gorilla/websocket"
)

type client struct {
	URL        string
	Conn       *websocket.Conn
	httpClient *http.Client
}

func (c *client) Listen(w io.Writer, fields []string, urlstr string) error {

	data, msgType, err := read(c.Conn)
	if err != nil {
		return err
	}
	if msgType == websocket.TextMessage {
		fmt.Fprint(w, "\n"+string(data))
	} else if msgType == websocket.BinaryMessage {
		req := serialize.DecodeRequest(data)
		fmt.Fprint(w, readRequestFields(fields, *req))
		req.URL, _ = url.Parse(urlstr)
		req.RequestURI = ""
		_, err := c.httpClient.Do(req)
		if err != nil {
			log.Fatalf("\ncli could not forwards message to local server, %v", err)
		}
	}
	return nil
}

func Newclient(serverURL string) *client {
	c := &client{}
	httpClient := &http.Client{}
	c.httpClient = httpClient
	c.Conn = NewConn(serverURL)
	c.URL = readURL(c.Conn)
	return c
}

func read(ws *websocket.Conn) ([]byte, int, error) {
	for {
		msgType, data, err := ws.ReadMessage()
		if err != nil {
			return []byte(""), -1, err
		}
		if len(data) != 0 {
			if msgType == websocket.TextMessage {
				return data, websocket.TextMessage, nil
			} else if msgType == websocket.BinaryMessage {
				return data, websocket.BinaryMessage, nil
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
	// listen to the server
	go func() {
		data, _, _ := read(ws)
		result <- string(data)
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
