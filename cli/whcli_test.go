package cli

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"whtester/serialize"

	"github.com/gorilla/websocket"
)

const fakeServerBaseURL = ":8080"
const fakeServerWSURL = "ws://localhost:8080/ws"

type localServerTestFake struct {
	req      http.Request
	received bool
	srv      *http.Server
}

func (l *localServerTestFake) Start() {
	fmt.Println(l.srv.ListenAndServe())
}

func (l *localServerTestFake) Close() {
	if err := l.srv.Shutdown(context.Background()); err != nil {
		log.Fatalf("shutting down local server tests fake: %s", err)
	}
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
	srv *http.Server
}

func (s *serverTestFake) Close() {
	s.ws.Close()
	err := s.srv.Shutdown(context.Background())
	if err != nil {
		log.Fatalf("shutting down server: %s", err)
	}
	fmt.Println("closing server")
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
	go func() {
		err := s.srv.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatalf("starting fake server: %s", err)
		}
	}()
	fmt.Println("server started")
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
	srv := &http.Server{
		Handler: mux,
		Addr:    fakeServerBaseURL,
	}
	s.mux = mux
	s.srv = srv
	return s
}

func TestWhCLI(t *testing.T) {

	// start server
	s := NewserverTestFake()
	go s.Start()
	defer s.Close()

	// wait for the server to start
	time.Sleep(time.Millisecond)

	t.Run("cli establishes websocket connection with the server", func(t *testing.T) {

		// try to connect to the server
		c := Newclient(fakeServerWSURL)
		defer c.Conn.Close()
		if c.Conn == nil {
			t.Fatal("cli didn't establish a connection with the server")
		}
	})

	t.Run("cli receives message from the server", func(t *testing.T) {

		buf := new(bytes.Buffer)

		// make clinet connection
		c := Newclient(fakeServerWSURL)
		defer c.Conn.Close()
		want := "this is a temp message"
		s.WriteMessage(want)
		c.Read(buf, nil, []int{5555})
		if buf.String() == "" {
			t.Error("expected a message to be writtem")
		}
	})

	t.Run("cli prints the same message, in a new line it receives from the server", func(t *testing.T) {

		buf := new(bytes.Buffer)

		c := Newclient(fakeServerWSURL)
		defer c.Conn.Close()

		msg := "message sent"
		s.WriteMessage(msg)
		c.Read(buf, nil, []int{5555})
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

		c := Newclient(fakeServerWSURL)
		defer c.Conn.Close()

		s.WriteEncodedRequest("this is a test")
		fields := []string{"Body", "Method", "URL", "Header"}

		c.Read(buf, fields, []int{5555})
		got := buf.String()
		for _, field := range fields {
			if !strings.Contains(got, field) {
				t.Errorf("output does not contain field %q, got %q", field, got)
			}
		}
	})

	t.Run("client forwards the received request to locally running server", func(t *testing.T) {
		// create a new client
		c := Newclient(fakeServerWSURL)
		defer c.Conn.Close()

		// create a new local server
		lsrv := NewLocalServerTestFake(":5555")
		go lsrv.Start()
		defer lsrv.Close()

		// write message to cli
		s.WriteEncodedRequest("this is a test")

		buf := new(bytes.Buffer)
		c.Read(buf, []string{"Body"}, []int{5555})

		// check if the local server received the message
		if lsrv.received == false {
			t.Errorf("local server didn't receive any request")
		}
	})
}

func TestForwardMultiplePorts(t *testing.T) {
	// create a new request
	body := strings.NewReader("hello world")
	req, err := http.NewRequest(http.MethodPost, "http://testurl.com", body)
	if err != nil {
		t.Errorf("creating new request: %s", err)
	}

	// create a new server
	srv := NewserverTestFake()
	go srv.Start()
	defer srv.Close()

	// wait for some time to get the server started
	time.Sleep(time.Millisecond)

	// create a new client
	c := Newclient(fakeServerWSURL)
	defer c.Conn.Close()

	// spin up new locally running server
	lsrv1 := NewLocalServerTestFake(":5556")
	go lsrv1.Start()
	defer lsrv1.Close()

	lsrv2 := NewLocalServerTestFake(":5557")
	go lsrv2.Start()
	defer lsrv2.Close()

	// wait for some time to get the server started
	time.Sleep(time.Millisecond)

	ports := []int{5556, 5557}
	reqblob := serialize.EncodeRequest(req)
	forwardRequestToPorts(c, reqblob, ports)
	if !lsrv1.received {
		t.Errorf("local server 1 didn't receive message")
	}
	if !lsrv2.received {
		t.Errorf("local server 2 didn't receive message")
	}
}

func BenchmarkReadRequest(b *testing.B) {
	// create a new http request
	body := bytes.NewBuffer([]byte("arbitary body for http request"))
	req := httptest.NewRequest(http.MethodPost, "http://benchmark:8080", body)
	fields := []string{"Header", "Method", "Body"}

	for i := 0; i < b.N; i++ {
		ReadRequestFields(fields, *req)
	}
}
