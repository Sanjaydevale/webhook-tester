package server

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"whtester/serialize"

	"github.com/gorilla/websocket"
)

var (
	PongWaitTime = 1 * time.Minute
	PingWaitTime = (PongWaitTime * 9) / 10
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type handler func(w http.ResponseWriter, r *http.Request)

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h(w, r)
}

type client struct {
	url string
	ws  *websocket.Conn
	handler
}

type Manager struct {
	ClientList map[string]client
	sync.RWMutex
}

func NewManager() *Manager {
	m := Manager{}
	m.ClientList = make(map[string]client)
	return &m
}

func (s *Manager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	subdomain := strings.Split(r.Host, ".")[0]
	client, ok := s.ClientList[subdomain]
	if ok {
		client.handler(w, r)
	} else {
		w.Write([]byte("client connection closed"))
	}
}

func (m *Manager) AddNewClient(u string, ws *websocket.Conn) {
	m.Lock()
	defer m.Unlock()
	uStruct, _ := url.Parse(u)
	// handle client conn
	newClient := &client{
		url: u,
		ws:  ws,
		handler: func(w http.ResponseWriter, r *http.Request) {
			//handle client
			if r.Method == http.MethodPost {
				msg := serialize.EncodeRequest(r)
				ws.WriteMessage(websocket.BinaryMessage, msg)
				w.WriteHeader(http.StatusAccepted)
			} else {
				w.WriteHeader(http.StatusForbidden)
			}
		},
	}
	subdomain := strings.Split(uStruct.Host, ".")[0]
	m.ClientList[subdomain] = *newClient
	go m.HandleClient(newClient)
}

func (m *Manager) RemoveClient(c *client) {
	m.Lock()
	defer m.Unlock()
	clientURL, _ := url.Parse(c.url)
	clientKey := strings.Split(clientURL.Host, ".")[0]
	fmt.Println("\nremoved client : ", c.url)
	c.ws.Close()
	delete(m.ClientList, clientKey)
}

func (m *Manager) HandleClient(c *client) {
	c.ws.SetReadDeadline(time.Now().Add(PongWaitTime))
	c.ws.SetPongHandler(func(appData string) error {
		return c.ws.SetReadDeadline(time.Now().Add(PongWaitTime))
	})
	defer m.RemoveClient(c)
	ticker := time.NewTicker(PingWaitTime)
	// send pings to client
	go func() {
		for {
			<-ticker.C
			c.ws.WriteMessage(websocket.PingMessage, []byte(""))
		}
	}()
	// read message from client to trigger pong handler
	for {
		_, _, err := c.ws.ReadMessage()
		// if err != nil connection is closed
		// remove the client on returning
		if err != nil {
			return
		}
	}
}

// Generates a random string and appends to the provided scheme and domain
func GenerateRandomURL(scheme string, domain string, subDomainLen int) string {
	var randSubDomain = GenerateRandomString(subDomainLen)
	u := url.URL{
		Scheme: scheme,
		Host:   strings.Join([]string{randSubDomain, domain}, "."),
		Path:   "",
	}
	return u.String()
}

func GenerateRandomString(strLen int) string {
	var charSet = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	var randStr = make([]rune, strLen)
	for i := 0; i < strLen; i++ {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(strLen)))
		if err != nil {
			log.Fatalf("unable to generate random number, %v", err)
		}
		randStr[i] = charSet[index.Int64()]
	}
	return string(randStr)
}

// Check if the given url is valid
func CheckValidURL(u string) bool {
	ustruct, err := url.ParseRequestURI(u)
	if err != nil {
		return false
	}
	if !(ustruct.Scheme == "http" || ustruct.Scheme == "https") || (ustruct.Host == "") {
		return false
	}
	return true
}

func NewWebHookHandler(clientsManager *Manager, domain string) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Fatalf("error establishing websocket connection")
		}
		u := []byte(GenerateRandomURL("http", domain, 8))
		clientsManager.AddNewClient(string(u), ws)
		fmt.Printf("\nnew client: %s", u)
		ws.WriteMessage(websocket.TextMessage, u)
	})
	mux.Handle("/", clientsManager)
	return mux
}
