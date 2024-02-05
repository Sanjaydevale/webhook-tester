package server

import (
	"math/rand"
	"net/url"
	"strings"
)

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
