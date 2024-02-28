package server

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
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
	})
}

// Generates a random string and appends to the provided scheme and domain
func GenerateRandomURL(scheme string, domain string, subDomainLen int) string {
	var charSet = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
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

func EncodeRequest(req *http.Request) []byte {
	buf := bytes.NewBuffer([]byte{})
	req.ParseForm()

	encoder := gob.NewEncoder(buf)
	encoder.Encode(req.Method)
	encoder.Encode(req.URL)
	encoder.Encode(req.Proto)
	encoder.Encode(req.ProtoMajor)
	encoder.Encode(req.ProtoMinor)
	encoder.Encode(req.Header)
	data, _ := io.ReadAll(req.Body)
	req.Body = io.NopCloser(bytes.NewBuffer(data))
	encoder.Encode(data)
	encoder.Encode(req.ContentLength)
	encoder.Encode(req.TransferEncoding)
	encoder.Encode(req.Host)
	encoder.Encode(req.Form)
	encoder.Encode(req.PostForm)
	encoder.Encode(req.Trailer)
	encoder.Encode(req.RemoteAddr)
	encoder.Encode(req.RequestURI)
	return buf.Bytes()
}

func DecodeRequest(buf []byte) *http.Request {
	req := http.Request{}
	decoder := gob.NewDecoder(bytes.NewBuffer(buf))
	decoder.Decode(&req.Method)
	decoder.Decode(&req.URL)
	decoder.Decode(&req.Proto)
	decoder.Decode(&req.ProtoMajor)
	decoder.Decode(&req.ProtoMinor)
	decoder.Decode(&req.Header)
	body := []byte{}
	decoder.Decode(&body)
	req.Body = io.NopCloser(bytes.NewBuffer(body))
	decoder.Decode(&req.ContentLength)
	decoder.Decode(&req.TransferEncoding)
	decoder.Decode(&req.Host)
	decoder.Decode(&req.Form)
	decoder.Decode(&req.PostForm)
	decoder.Decode(&req.Trailer)
	decoder.Decode(&req.RemoteAddr)
	decoder.Decode(&req.RequestURI)
	return &req
}
