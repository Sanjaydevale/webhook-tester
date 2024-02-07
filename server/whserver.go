package server

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
)

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

func (s *Subdomains) AddNewClient(u string) {
	uStruct, _ := url.Parse(u)
	(*s)[uStruct.Host] = handler(func(w http.ResponseWriter, r *http.Request) {
		//handle client
		//forward any post request to cli
		fmt.Println("hello this is getting called")
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
