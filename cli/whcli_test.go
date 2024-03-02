package cli_test

import (
	"bytes"
	"net/http"
	"testing"
	"whtester/cli"

	"github.com/gorilla/websocket"
)

type serverTestFake struct {
	ws  *websocket.Conn
	mux *http.ServeMux
}

func (s serverTestFake) WriteMessage(msg string) {
	s.ws.WriteMessage(websocket.TextMessage, []byte(msg))
}

func (s *serverTestFake) Start() {
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
		c := cli.Newclient()
		defer c.Conn.Close()
		if c.Conn == nil {
			t.Fatal("cli didn't establish a connection with the server")
		}
	})

	t.Run("cli receives message from the server", func(t *testing.T) {

		buf := new(bytes.Buffer)

		// make clinet connection
		c := cli.Newclient()
		defer c.Conn.Close()
		want := "this is a temp message"
		s.WriteMessage(want)
		c.PrintMessage(buf)
		if buf.String() == "" {
			t.Error("expected a message to be writtem")
		}
	})

	t.Run("cli prints the same, it receives from the server", func(t *testing.T) {

		buf := new(bytes.Buffer)

		c := cli.Newclient()
		defer c.Conn.Close()

		want := "message sent"
		s.WriteMessage(want)
		c.PrintMessage(buf)

		if buf.String() != want {
			t.Errorf("got %q, want %q", buf.String(), want)
		}
	})
}
