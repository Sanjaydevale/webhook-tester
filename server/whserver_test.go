package server

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

type clientTestFake struct {
	output map[int][][]byte
	url    string
	key    string
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

func AddClientToGroup(t testing.TB, groupUrl string, key string) *clientTestFake {
	t.Helper()

	header := make(http.Header)
	header.Set("url", groupUrl)
	header.Set("key", key)

	c := &clientTestFake{}
	c.client = http.Client{}
	c.domain = "ws://localhost:8080/wsold"
	c.key = key
	c.url = groupUrl
	c.output = make(map[int][][]byte)
	ws, _, err := websocket.DefaultDialer.Dial(c.domain, header)
	c.ws = ws
	if err != nil {
		log.Fatalf("error in clientTestFake establishing connection with whserver, %v", err.Error())
	}
	return c
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
		// message from server contains both url and password
		url := strings.Split(string(p), "\n")[0]
		key := strings.Split(string(p), "password: ")[1]
		if url != "" && key != "" {
			c.url = url
			c.key = key
			break
		}
	}
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
		rawURL := GenerateRandomURL("http", "localhost:8080", 8)
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
		url := GenerateRandomURL("http", "localhost:8080", 8)
		if !CheckValidURL(url) {
			t.Errorf("generates an invalid url got %s", url)
		}
	})
	t.Run("generates a random URL everytime", func(t *testing.T) {
		var urlList []string
		for i := 0; i < 10; i++ {
			url := GenerateRandomURL("http", "localhost:8080", 8)
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
	_, close := startServer(t)
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

func TestHandlingMultipleClients(t *testing.T) {
	// start server
	mgr, close := startServer(t)
	defer close()

	// wait for the server to start
	time.Sleep(1 * time.Second)

	t.Run("can add a new client to an exisiting group", func(t *testing.T) {
		// create a new client
		c1 := NewclientTestFake()
		defer c1.ws.Close()

		// add a new client to an exisitng group
		c2 := AddClientToGroup(t, c1.url, c1.key)
		defer c2.ws.Close()

		clientURL, _ := url.Parse(c1.url)
		clientKey := strings.Split(clientURL.Host, ".")[0]
		assert.Equal(t, 2, len(mgr.ClientList[clientKey].clients))
	})

	t.Run("group is remove, if all clients are disconnected", func(t *testing.T) {
		// create new clients
		c1 := NewclientTestFake()
		c2 := AddClientToGroup(t, c1.url, c1.key)

		// disconnect clients
		c1.ws.Close()
		c2.ws.Close()

		// wait for the server to remove the clients
		time.Sleep(time.Millisecond)

		clientURL, _ := url.Parse(c1.url)
		clientKey := strings.Split(clientURL.Host, ".")[0]
		_, ok := mgr.ClientList[clientKey]
		assert.False(t, ok)
	})

	t.Run("message sent to a group is received by all the clients", func(t *testing.T) {
		// create a new client
		c1 := NewclientTestFake()
		defer c1.ws.Close()

		// add a new client to an exisitng group
		c2 := AddClientToGroup(t, c1.url, c1.key)
		defer c2.ws.Close()

		// add a new client to an exisitng group
		c3 := AddClientToGroup(t, c1.url, c1.key)
		defer c3.ws.Close()

		// stream data from server
		go c1.read()
		go c2.read()
		go c3.read()

		whTrigger := webhookTrigger{
			url: c1.url,
		}
		whTrigger.sendPOSTRequest()

		time.Sleep(time.Millisecond)
		// check if the binary message (request) by all the
		// clients is equal
		c1out := c1.output[websocket.BinaryMessage][0]
		c2out := c2.output[websocket.BinaryMessage][0]
		c3out := c3.output[websocket.BinaryMessage][0]

		assert.Equal(t, c1out, c2out)
		assert.Equal(t, c1out, c2out)

		assert.Equal(t, c2out, c1out)
		assert.Equal(t, c2out, c3out)

		assert.Equal(t, c3out, c1out)
		assert.Equal(t, c3out, c2out)
	})
}

func TestClienConnection(t *testing.T) {

	// set PingWait and PongWait for server
	PongWaitTime = 1 * time.Second
	PingWaitTime = (PongWaitTime * 9) / 10

	// start server
	manager := NewManager()
	mux := NewWebHookHandler(manager, "localhost:8080")
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
			time.Sleep((PongWaitTime * 13) / 10)

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
			time.Sleep(PongWaitTime)

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
			time.Sleep(PongWaitTime)

			// check if client connection is open
			checkClientConnectionOpen(t, manager, c)
		})

		t.Run("server removes client if client closes websocket connection", func(t *testing.T) {
			t.Parallel()
			c := NewclientTestFake()
			c.ws.Close()
			time.Sleep(PongWaitTime)
			checkClientConnectionClose(t, manager, c)
		})
	})
}

func checkClientConnectionClose(t testing.TB, manager *Manager, c *clientTestFake) {
	u, _ := url.Parse(c.url)
	subdomain := strings.Split(u.Host, ".")[0]
	_, ok := manager.ClientList[subdomain]
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

func checkClientConnectionOpen(t testing.TB, manager *Manager, c *clientTestFake) {
	u, _ := url.Parse(c.url)
	subdomain := strings.Split(u.Host, ".")[0]
	_, ok := manager.ClientList[subdomain]
	if ok == false {
		t.Fatal("client is not present in manager")
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

func startServer(t testing.TB) (*Manager, func() error) {
	t.Helper()
	clientManager := NewManager()
	mux := NewWebHookHandler(clientManager, "localhost:8080")
	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	go func() {
		fmt.Println("\n", srv.ListenAndServe())
	}()
	return clientManager, srv.Close
}
