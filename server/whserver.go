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

	"github.com/google/uuid"
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
	uid string
	handler
}

type clientGroup struct {
	clients map[string]client
}

type Manager struct {
	ClientList map[string]clientGroup
	Passwords  map[string]string
	sync.RWMutex
}

func NewManager() *Manager {
	m := Manager{}
	m.ClientList = make(map[string]clientGroup)
	m.Passwords = make(map[string]string)
	return &m
}

func (s *Manager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	subdomain := strings.Split(r.Host, ".")[0]
	clientGroup, ok := s.ClientList[subdomain]
	for _, client := range clientGroup.clients {
		if ok {
			client.handler(w, r)
		} else {
			w.Write([]byte("client connection closed"))
		}
	}
}

func (m *Manager) AddNewClient(u string, ws *websocket.Conn) {
	m.Lock()
	defer m.Unlock()
	uStruct, _ := url.Parse(u)
	uid := uuid.New().String()
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
		uid: uid,
	}
	subdomain := strings.Split(uStruct.Host, ".")[0]
	group, ok := m.ClientList[subdomain]
	if !ok {
		newGroup := &clientGroup{
			clients: make(map[string]client),
		}
		m.ClientList[subdomain] = *newGroup
		newGroup.clients[uid] = *newClient
	} else {
		group.clients[uid] = *newClient
	}
	fmt.Printf("\nnew client: %s", uid)
	go m.HandleClient(newClient)
}

func (m *Manager) RemoveClient(c *client) {
	m.Lock()
	defer m.Unlock()
	clientURL, _ := url.Parse(c.url)
	clientKey := strings.Split(clientURL.Host, ".")[0]
	c.ws.Close()
	// delete client from the group
	group, ok := m.ClientList[clientKey]
	delete(group.clients, c.uid)
	fmt.Printf("\nremoved client : %s", c.uid)

	// no client in the group delete the group
	if ok && len(group.clients) == 0 {
		delete(m.ClientList, clientKey)
		// delete the client also from passwords
		delete(m.Passwords, clientKey)
		fmt.Printf("\nremove client group: %s", clientKey)
	}
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
		// generate random password
		password := GenerateRandomString(6)
		clientsManager.Passwords[string(u)] = password
		clientsManager.AddNewClient(string(u), ws)
		// send password and unique url to the client
		u = append(u, []byte(fmt.Sprintf("\npassword: %s", password))...)
		ws.WriteMessage(websocket.TextMessage, u)
	})

	mux.HandleFunc("/wsold", func(w http.ResponseWriter, r *http.Request) {
		// upgrade connection to websockets
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Fatalf("error establishing websocket connection")
		}

		Url := r.Header.Get("url")
		Key := r.Header.Get("key")

		if err != nil {
			ws.WriteMessage(websocket.TextMessage, []byte("invalid json data"))
			ws.Close()
			return
		}

		// check if the group exists
		key, ok := clientsManager.Passwords[Url]
		if !ok {
			ws.WriteMessage(websocket.TextMessage, []byte("invalid group, grop does not exisit"))
			ws.Close()
			return
		}

		if key == Key {
			clientsManager.AddNewClient(Url, ws)
		}
	})
	mux.Handle("/", clientsManager)
	return mux
}
