package server_test

import (
	"strings"
	"testing"
	"whtester/cmd/server"
)

func TestRandomURL(t *testing.T) {

	//assumed generated url format
	//https://Y38IpO3Ow4.example.com
	t.Run("genereates a random URL with 8 character length subdomain", func(t *testing.T) {
		url := server.GenerateRandomURL("localhost:8080", 8)
		url = strings.TrimLeft(url, "htps:/")
		subdomain := strings.Split(url, ".")[0]
		want := 8
		if len(subdomain) != want {
			t.Errorf("invlaid subdomain length, got %d, want %d", len(subdomain), want)
		}
	})

	t.Run("generates a random URL everytime", func(t *testing.T) {
		var urlList []string
		for i := 0; i < 10; i++ {
			url := server.GenerateRandomURL("localhost:8080", 8)
			for _, u := range urlList {
				if u == url {
					t.Fatalf("found duplicate urls, %s", u)
				}
			}
			urlList = append(urlList, url)
		}
	})
}
