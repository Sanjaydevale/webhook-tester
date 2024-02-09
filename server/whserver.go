package server

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"

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
		ws.WriteMessage(websocket.TextMessage, []byte("hello from server on post request"))
		w.Write([]byte(fmt.Sprintf("server is listening at %s", r.Host)))
	})
}

func GenerateRandomURL(scheme string, domain string, subDomainLen int) string {
	var charSet = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	var randSubDomain []rune
	for i := 0; i < subDomainLen; i++ {
		randSubDomain = append(randSubDomain, charSet[rand.Int()%len(charSet)])
	}
	u := url.URL{
		Scheme: scheme,
		Host:   strings.Join([]string{string(randSubDomain), domain}, "."),
		Path:   "",
	}
	return u.String()
}

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
		ws.WriteMessage(websocket.TextMessage, u)
	})
	mux.Handle("/", subDomainsHandler)
	return mux
}
