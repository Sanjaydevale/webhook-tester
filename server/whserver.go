package server

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"whtester/serialize"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type handler func(w http.ResponseWriter, r *http.Request)

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h(w, r)
}

type Subdomains map[string]http.Handler

func (s Subdomains) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler := s[r.Host]; handler != nil {
		handler.ServeHTTP(w, r)
	}
}

func (s *Subdomains) AddNewClient(u string, ws *websocket.Conn) {
	uStruct, _ := url.Parse(u)
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
}

// Generates a random string and appends to the provided scheme and domain
func GenerateRandomURL(scheme string, domain string, subDomainLen int) string {
	var charSet = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	var randSubDomain []rune
	for len(randSubDomain) != subDomainLen {
		randSubDomain = append(randSubDomain, charSet[rand.Int()%len(charSet)])
	}
	u := url.URL{
		Scheme: scheme,
		Host:   strings.Join([]string{string(randSubDomain), domain}, "."),
		Path:   "",
	}
	return u.String()
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

func NewWebHookHandler() *http.ServeMux {
	subDomainsHandler := Subdomains{}
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Fatalf("error establishing websocket connection")
		}
		u := []byte(GenerateRandomURL("http", "localhost:8080", 8))
		subDomainsHandler.AddNewClient(string(u), ws)
		fmt.Printf("\nnew client: %s", u)
		ws.WriteMessage(websocket.TextMessage, u)
	})
	mux.Handle("/", subDomainsHandler)
	return mux
}
