package cli_test

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"whtester/cli"
	"whtester/serialize"

	"github.com/gorilla/websocket"
)

type localServerTestFake struct {
	req      http.Request
	received bool
	srv      *http.Server
}

func (l *localServerTestFake) Start() {
	fmt.Println(l.srv.ListenAndServe())
}

func (l *localServerTestFake) Close() {
	l.srv.Close()
}

func (l *localServerTestFake) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l.req = *r
	l.received = true
}

func NewLocalServerTestFake(port string) *localServerTestFake {
	lsrv := &localServerTestFake{
		received: false,
	}
	srv := &http.Server{Addr: port, Handler: lsrv}
	lsrv.srv = srv
	return lsrv
}

type serverTestFake struct {
	ws  *websocket.Conn
	mux *http.ServeMux
}

func (s serverTestFake) WriteMessage(msg string) {
	s.ws.WriteMessage(websocket.TextMessage, []byte(msg))
}

func (s serverTestFake) WriteEncodedRequest(body string) {
	b := bytes.NewBuffer([]byte(body))
	req, _ := http.NewRequest(http.MethodPost, "tempurl", b)
	msg := serialize.EncodeRequest(req)
	s.ws.WriteMessage(websocket.BinaryMessage, msg)
}

func (s *serverTestFake) Start() {
	fmt.Println("server started")
	http.ListenAndServe(":8080", s.mux)
}

func NewserverTestFake() *serverTestFake {
	s := &serverTestFake{}
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		ws, _ := upgrader.Upgrade(w, r, nil)
		s.ws = ws
		ws.WriteMessage(websocket.TextMessage, []byte("tempURL"))
	})
	s.mux = mux
	return s
}

func TestWhCLI(t *testing.T) {

	// start server
	s := NewserverTestFake()
	go s.Start()
	t.Run("cli establishes websocket connection with the server", func(t *testing.T) {

		// try to connect to the server
		c := cli.Newclient("ws://localhost:8080/ws")
		defer c.Conn.Close()
		if c.Conn == nil {
			t.Fatal("cli didn't establish a connection with the server")
		}
	})

	t.Run("cli receives message from the server", func(t *testing.T) {

		buf := new(bytes.Buffer)

		// make clinet connection
		c := cli.Newclient("ws://localhost:8080/ws")
		defer c.Conn.Close()
		want := "this is a temp message"
		s.WriteMessage(want)
		c.Listen(buf, nil, "http://localhost:5555")
		if buf.String() == "" {
			t.Error("expected a message to be writtem")
		}
	})

	t.Run("cli prints the same message, in a new line it receives from the server", func(t *testing.T) {

		buf := new(bytes.Buffer)

		c := cli.Newclient("ws://localhost:8080/ws")
		defer c.Conn.Close()

		msg := "message sent"
		s.WriteMessage(msg)
		c.Listen(buf, nil, "")
		want := "\n" + msg
		if buf.String() != want {
			t.Errorf("got %q, want %q", buf.String(), want)
		}
	})

	t.Run("cli prints only the specified fields of the request", func(t *testing.T) {

		// create a new local server
		lsrv := NewLocalServerTestFake(":5555")
		go lsrv.Start()
		defer lsrv.Close()

		buf := new(bytes.Buffer)

		c := cli.Newclient("ws://localhost:8080/ws")
		defer c.Conn.Close()

		s.WriteEncodedRequest("this is a test")
		fields := []string{"Body", "Method", "URL", "Header"}

		c.Listen(buf, fields, "http://localhost:5555")
		got := buf.String()
		for _, field := range fields {
			if !strings.Contains(got, field) {
				t.Errorf("output does not contain field %q, got %q", field, got)
			}
		}
	})

	t.Run("client forwards the received request to locally running server", func(t *testing.T) {
		// create a new client
		c := cli.Newclient("ws://localhost:8080/ws")
		defer c.Conn.Close()

		// create a new local server
		lsrv := NewLocalServerTestFake(":5555")
		go lsrv.Start()
		defer lsrv.Close()

		// write message to cli
		s.WriteEncodedRequest("this is a test")

		buf := new(bytes.Buffer)
		c.Listen(buf, []string{"Body"}, "http://localhost:5555")

		// check if the local server received the message
		if lsrv.received == false {
			t.Errorf("local server didn't receive any request")
		}
	})
}
