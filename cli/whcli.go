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

type Client struct {
	URL        string
	Key        string
	Conn       *websocket.Conn
	httpClient *http.Client
}

var AvailabeFields = map[string]struct{}{
	"Method": {}, "URL": {}, "Proto": {}, "ProtoMajor": {}, "ProtoMinor": {},
	"Header": {}, "Body": {}, "ContentLength": {}, "TransferEncoding": {}, "Close": {},
	"Host": {}, "RemoteAddr": {}, "RequestURI": {},
}

func (c *Client) Stream(w io.Writer, fields []string, ports []int) {
	for {
		c.Read(w, fields, ports)
	}
}

func (c *Client) Read(w io.Writer, fields []string, ports []int) {
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
		forwardRequestToPorts(c, data, ports)
	}
}

func forwardRequestToPorts(c *Client, reqblob []byte, ports []int) {
	for _, port := range ports {
		req := serialize.DecodeRequest(reqblob)
		forwardRequest(c, req, port)
	}
}

func forwardRequest(c *Client, req *http.Request, port int) {
	req.URL, _ = url.Parse(fmt.Sprintf("http://localhost:%d", port))
	req.RequestURI = ""
	_, err := c.httpClient.Do(req)
	if err != nil {
		log.Fatalf("\ncli could not forwards message to local server, %v", err)
	}
}

func Newclient(serverURL string) *Client {
	c := &Client{}
	httpClient := &http.Client{}
	c.httpClient = httpClient
	c.Conn = NewConn(serverURL)
	readURLAndKey(c)
	return c
}

func ConnToGroup(serverURL string, groupURL string, key string) *Client {
	c := &Client{}
	httpClient := &http.Client{}
	c.httpClient = httpClient
	c.URL = groupURL
	c.Key = key
	c.Conn = NewConnGroup(serverURL, groupURL, key)
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

func readURLAndKey(c *Client) {
	done := make(chan bool)
	// listen to the server
	go func() {
		msgType, data, err := c.Conn.ReadMessage()
		if err != nil {
			log.Fatalf("error reading URL from server, %v", err)
		}
		if msgType != websocket.TextMessage {
			log.Fatalf("expected to received URL from server, go message of type %d", msgType)
		}
		url := strings.Split(string(data), "\n")[0]
		key := strings.Split(string(data), "password: ")[1]
		c.URL = url
		c.Key = key
		done <- true
	}()

	select {
	case <-done:
		return
	case <-time.After(5 * time.Second):
		log.Fatalf("took too long to read message from server")
	}

}

func NewConnGroup(wsLink string, url string, key string) *websocket.Conn {
	header := make(http.Header)
	header.Set("url", url)
	header.Set("key", key)
	dailer := websocket.DefaultDialer
	dailer.HandshakeTimeout = time.Minute
	ws, _, err := dailer.Dial(wsLink, header)
	if err != nil {
		log.Fatalf("error establishing websocket connection: %v", err.Error())
	}
	return ws
}

func NewConn(wsLink string) *websocket.Conn {
	websocket.DefaultDialer.HandshakeTimeout = time.Minute
	dailer := websocket.DefaultDialer
	dailer.HandshakeTimeout = time.Minute
	ws, _, err := dailer.Dial(wsLink, nil)
	if err != nil {
		log.Fatalf("error establishing websocket connection: %v", err.Error())
	}
	return ws
}
