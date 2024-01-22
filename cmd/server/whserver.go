package server

import (
	"math/rand"
	"strings"
)

func GenerateRandomURL(domain string, subDomainLen int) string {
	var charSet = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	var randSubDomain []rune
	for i := 0; i < subDomainLen; i++ {
		randSubDomain = append(randSubDomain, charSet[rand.Int()%len(charSet)])
	}
	return strings.Join([]string{string(randSubDomain), string(domain)}, ".")
}
