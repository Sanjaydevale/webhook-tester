package server

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"strings"
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

type Manager map[string]http.Handler

func (s Manager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler := s[r.Host]; handler != nil {
		handler.ServeHTTP(w, r)
	}
}

func (s *Manager) AddNewClient(u string, ws *websocket.Conn) {
	uStruct, _ := url.Parse(u)
	// handle client conn
	(*s)[uStruct.Host] = handler(func(w http.ResponseWriter, r *http.Request) {
		//handle client
		if r.Method == http.MethodPost {
			msg := serialize.EncodeRequest(r)
			ws.WriteMessage(websocket.BinaryMessage, msg)
			w.WriteHeader(http.StatusAccepted)
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
	})
	go s.HandleClient(u, ws)
}

func (m *Manager) RemoveClient(u string, ws *websocket.Conn) {
	clientURL, _ := url.Parse(u)
	clientKey := clientURL.Host
	fmt.Println("\nremoved client : ", u)
	ws.Close()
	delete(*m, clientKey)
}

func (m *Manager) HandleClient(u string, ws *websocket.Conn) {
	ws.SetReadDeadline(time.Now().Add(PongWaitTime))
	defer m.RemoveClient(u, ws)
	ticker := time.NewTicker(PingWaitTime)
	// send pings to client
	go func() {
		for {
			<-ticker.C
			ws.WriteMessage(websocket.PingMessage, []byte(""))
		}
	}()
	// read message from client to trigger pong handler
	for {
		_, _, err := ws.ReadMessage()
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

func NewWebHookHandler(clientsManager Manager) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Fatalf("error establishing websocket connection")
		}
		u := []byte(GenerateRandomURL("http", "localhost:8080", 8))
		clientsManager.AddNewClient(string(u), ws)
		fmt.Printf("\nnew client: %s", u)
		ws.WriteMessage(websocket.TextMessage, u)
	})
	mux.Handle("/", clientsManager)
	return mux
}
