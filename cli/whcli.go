package cli

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
	"whtester/serialize"

	"github.com/gorilla/websocket"
)

type client struct {
	URL        string
	Conn       *websocket.Conn
	httpClient *http.Client
}

var AvailabeFields = map[string]struct{}{
	"Method": {}, "URL": {}, "Proto": {}, "ProtoMajor": {}, "ProtoMinor": {},
	"Header": {}, "Body": {}, "ContentLength": {}, "TransferEncoding": {}, "Close": {},
	"Host": {}, "RemoteAddr": {}, "RequestURI": {},
}

func (c *client) Stream(w io.Writer, fields []string, ports []int) {
	for {
		c.Read(w, fields, ports)
	}
}

func (c *client) Read(w io.Writer, fields []string, ports []int) {
	msgType, data, err := c.Conn.ReadMessage()
	if err != nil {
		log.Fatalf("\nerror reading message from server, %v\n", err)
		return
	}
	if msgType == websocket.TextMessage {
		fmt.Fprint(w, "\n"+string(data))
	} else if msgType == websocket.BinaryMessage {
		// client recevied encoded HTTP POST request
		// decode binary  blob into HTTP request struct
		req := serialize.DecodeRequest(data)

		// print the specified fields
		fmt.Fprint(w, ReadRequestFields(fields, *req))

		// forward request to locally running program
		forwardRequestToPorts(c, req, ports)
	}
}

func forwardRequestToPorts(c *client, req *http.Request, ports []int) {
	for _, port := range ports {
		fmt.Println(port)
		forwardRequest(c, req, port)
	}
}

func forwardRequest(c *client, req *http.Request, port int) {
	req.URL, _ = url.Parse(fmt.Sprintf("http://localhost:%d", port))
	req.RequestURI = ""
	_, err := c.httpClient.Do(req)
	if err != nil {
		log.Fatalf("\ncli could not forwards message to local server, %v", err)
	}
}

func Newclient(serverURL string) *client {
	c := &client{}
	httpClient := &http.Client{}
	c.httpClient = httpClient
	c.Conn = NewConn(serverURL)
	c.URL = readURL(c.Conn)
	return c
}

// reads only specifies HTTP request fields and returns them in string format
func ReadRequestFields(fields []string, req http.Request) string {
	var out strings.Builder
	r := reflect.ValueOf(req)
	for _, f := range fields {
		if r.FieldByName(f) == reflect.ValueOf(nil) {
			fmt.Printf("\ndoes not have field, %s\n", f)
			fmt.Println("available fields:")
			for i := 0; i < r.NumField(); i++ {
				fmt.Println(r.Type().Field(i).Name)
			}
			continue
		}
		field := fmt.Sprintf("\n%s :%v", f, r.FieldByName(f).Interface())
		out.WriteString(field)
		out.WriteString("\n")
	}
	return out.String()
}

func readURL(ws *websocket.Conn) string {
	result := make(chan string, 1)
	// listen to the server
	go func() {
		msgType, data, err := ws.ReadMessage()
		if err != nil {
			log.Fatalf("error reading URL from server, %v", err)
		}
		if msgType != websocket.TextMessage {
			log.Fatalf("expected to received URL from server, go message of type %d", msgType)
		}
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
