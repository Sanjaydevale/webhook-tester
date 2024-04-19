package server_test

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
	"whtester/server"

	"github.com/gorilla/websocket"
)

type clientTestFake struct {
	output map[int][][]byte
	url    string
	domain string
	ws     *websocket.Conn
	client http.Client
}

func (c *clientTestFake) read() {
	for {
		msgType, p, err := c.ws.ReadMessage()
		if err != nil {
			return
		}
		if string(p) != "" {
			if msgType == websocket.TextMessage {
				c.output[websocket.TextMessage] = append(c.output[websocket.TextMessage], p)
			} else if msgType == websocket.BinaryMessage {
				c.output[websocket.BinaryMessage] = append(c.output[websocket.BinaryMessage], p)
			}
		}
	}
}

func NewclientTestFake() *clientTestFake {
	c := &clientTestFake{}
	c.client = http.Client{}
	c.domain = "ws://localhost:8080/ws"
	ws, _, err := websocket.DefaultDialer.Dial(c.domain, nil)
	c.output = make(map[int][][]byte)
	//fmt.Println(c.output)
	c.ws = ws
	if err != nil {
		log.Fatalf("error in clientTestFake establishing connection with whserver, %v", err.Error())
	}

	//read url
	for {
		_, p, err := c.ws.ReadMessage()
		if err != nil {
			continue
		}
		if string(p) != "" {
			c.url = string(p)
			break
		}
	}
	c.ws = ws
	return c
}

type webhookTrigger struct {
	url string
}

func (w webhookTrigger) sendPOSTRequest() {
	body := bytes.NewReader([]byte("hello world, here are the changes made in the system"))
	_, err := http.Post(w.url, "applications/json", body)
	if err != nil {
		log.Fatalf("error making post request to the server, %v", err.Error())
	}
}

func (w webhookTrigger) sentGETRequest() *http.Response {
	resp, err := http.Get(w.url)
	if err != nil {
		log.Fatalf("error making get request to the server, %v", err.Error())
	}
	return resp
}

func TestRandomURL(t *testing.T) {

	//assumed generated url format
	//https://Y38IpO3Ow4.example.com
	t.Run("genereates a URL with 8 character length subdomain", func(t *testing.T) {
		rawURL := server.GenerateRandomURL("http", "localhost:8080", 8)
		paresedurl, err := url.Parse(rawURL)
		if err != nil {
			t.Errorf("unable to parse rawURL, got %s, error : %v", rawURL, err)
		}
		subdomain := strings.Split(paresedurl.Host, ".")[0]
		want := 8
		if len(subdomain) != want {
			t.Errorf("invlaid subdomain length, got %d(%s), want %d", len(subdomain), subdomain, want)
		}
	})

	t.Run("generate a valid URL", func(t *testing.T) {
		url := server.GenerateRandomURL("http", "localhost:8080", 8)
		if !server.CheckValidURL(url) {
			t.Errorf("generates an invalid url got %s", url)
		}
	})
	t.Run("generates a random URL everytime", func(t *testing.T) {
		var urlList []string
		for i := 0; i < 10; i++ {
			url := server.GenerateRandomURL("http", "localhost:8080", 8)
			for _, u := range urlList {
				if u == url {
					t.Fatalf("found duplicate urls, %s", u)
				}
			}
			urlList = append(urlList, url)
		}
	})
}

func TestForwardingMessage(t *testing.T) {
	// start server
	close := startServer(t)
	defer close()

	// wait for the server to start
	time.Sleep(1 * time.Second)

	t.Run("server pings the client on POST request to temp URL", func(t *testing.T) {

		// create a new client
		c := NewclientTestFake()
		defer c.ws.Close()

		// listend to the message from the server
		go func() {
			c.read()
		}()

		// make post request to the server
		whTrigger := webhookTrigger{
			url: c.url,
		}
		whTrigger.sendPOSTRequest()

		if len(c.output) == 0 {
			t.Error("expected POST request to be forwarded by the server, but got none")
		}
	})

	t.Run("server forwards only post request", func(t *testing.T) {
		// create a new client
		c := NewclientTestFake()
		defer c.ws.Close()

		// listend to the message from the server
		go func() {
			c.read()
		}()

		// make post request to the server
		whTrigger := webhookTrigger{
			url: c.url,
		}
		resp := whTrigger.sentGETRequest()
		if resp.StatusCode == http.StatusAccepted {
			t.Errorf("get requets not ignored got statusAccepted, %d", resp.StatusCode)
		}
	})
	t.Run("server sends the post request it receives in binary format to client", func(t *testing.T) {

		// create a new clent
		c := NewclientTestFake()
		defer c.ws.Close()

		// listend to the message from the server
		go func() {
			c.read()
		}()

		// create post requst
		data := []byte("this is the body of the new post request")
		body := bytes.NewBuffer(data)
		req, err := http.NewRequest(http.MethodPost, c.url, body)
		if err != nil {
			t.Errorf("error creating request, %v", err)
		}

		// make post request
		c.client.Do(req)

		// check if the data received by the client
		if len(c.output[websocket.BinaryMessage]) == 0 {
			t.Fatalf("didn't receive any binary message")
		}
	})
}

func TestClienConnection(t *testing.T) {

	// set PingWait and PongWait for server
	server.PongWaitTime = 1 * time.Second
	server.PingWaitTime = (server.PongWaitTime * 9) / 10

	// start server
	manager := server.NewManager()
	mux := server.NewWebHookHandler(manager, "localhost:8080")
	srv := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	go srv.ListenAndServe()
	defer srv.Close()

	t.Run("parallel test group", func(t *testing.T) {
		t.Run("server listens on generated URL", func(t *testing.T) {
			t.Parallel()
			c := NewclientTestFake()
			defer c.ws.Close()

			u := c.url
			res, err := http.Head(u)
			if err != nil {
				t.Errorf("error making http HEAD, %s", err.Error())
			}
			if res.StatusCode == 404 {
				t.Errorf("server is not listenting on URL, %s", u)
			}
		})

		t.Run("server sends ping messages to the client", func(t *testing.T) {
			t.Parallel()
			received := make(chan bool)
			// make a new client connection
			c := NewclientTestFake()
			c.ws.SetPingHandler(func(appData string) error {
				received <- true
				return nil
			})
			go c.read()
			select {
			case <-received:
				return
			case <-time.After(2 * time.Second):
				t.Fatal("didn't receive any pong messages from server")
			}
		})

		t.Run("disconnected client connection is deleted from Manager", func(t *testing.T) {
			t.Parallel()
			// create a new client connection
			c := NewclientTestFake()
			// client does not replay to ping's from server
			c.ws.SetPingHandler(func(appData string) error {
				return nil
			})

			// wait for pong timeout to occure in server
			time.Sleep((server.PongWaitTime * 13) / 10)

			// check if client still exists is Manager map
			u, _ := url.Parse(c.url)
			_, ok := manager.ClientList[u.Host]
			if ok == true {
				t.Fatal("client is present in manager")
			}

			// check if websocket connection is closed
			checkClientConnectionClose(t, manager, c)

		})

		t.Run("server removes client if, client closes websocket connection", func(t *testing.T) {
			t.Parallel()

			c := NewclientTestFake()
			c.ws.Close()

			// wait for pong time out
			time.Sleep(server.PongWaitTime)

			// check if client still exists is Manager map
			checkClientConnectionClose(t, manager, c)
		})

		t.Run("clients which pong's server, maintains connection", func(t *testing.T) {
			t.Parallel()
			// create a new client
			c := NewclientTestFake()
			// default ping handler of client sends pong as replay to ping
			go c.read()

			// wait for pong time out
			time.Sleep(server.PongWaitTime)

			// check if client connection is open
			checkClientConnectionOpen(t, manager, c)
		})

		t.Run("server removes client if client closes websocket connection", func(t *testing.T) {
			t.Parallel()
			c := NewclientTestFake()
			c.ws.Close()
			time.Sleep(server.PongWaitTime)
			checkClientConnectionClose(t, manager, c)
		})
	})
}

func checkClientConnectionClose(t testing.TB, manager *server.Manager, c *clientTestFake) {
	u, _ := url.Parse(c.url)
	_, ok := manager.ClientList[u.Host]
	if ok == true {
		t.Fatal("client is present in manager")
	}

	// check if websocket connection is closed
	var err error
	var closed = make(chan bool)
	go func() {
		_, _, err = c.ws.ReadMessage()
		if err != nil {
			closed <- true
		}
	}()

	select {
	case <-closed:
		return
	case <-time.After(3 * time.Second):
		t.Fatal("websocket connection is not closed")
	}
}

func checkClientConnectionOpen(t testing.TB, manager *server.Manager, c *clientTestFake) {
	u, _ := url.Parse(c.url)
	_, ok := manager.ClientList[u.Host]
	if ok == false {
		t.Fatal("client is present in manager")
	}

	// check if websocket connection is closed
	var err error
	var closed = make(chan bool)
	go func() {
		_, _, err = c.ws.ReadMessage()
		if err != nil {
			closed <- true
		}
	}()

	select {
	case <-closed:
		t.Fatalf("client connection is closed")
	case <-time.After(3 * time.Second):
		return
	}
}

func startServer(t testing.TB) func() error {
	t.Helper()
	clientManager := server.NewManager()
	mux := server.NewWebHookHandler(clientManager, "localhost:8080")
	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	go func() {
		fmt.Println(srv.ListenAndServe())
	}()
	return srv.Close
}
