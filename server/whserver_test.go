package server_test

import (
	"net/http"
	"strings"
	"testing"
	"whtester/cli"
	"whtester/server"
)

func TestRandomURL(t *testing.T) {

	//assumed generated url format
	//https://Y38IpO3Ow4.example.com
	t.Run("genereates a random URL with 8 character length subdomain", func(t *testing.T) {
		url := server.GenerateRandomURL("http", "localhost:8080", 8)
		url = strings.TrimLeft(url, "htps:/")
		subdomain := strings.Split(url, ".")[0]
		want := 8
		if len(subdomain) != want {
			t.Errorf("invlaid subdomain length, got %d, want %d", len(subdomain), want)
		}
	})

	t.Run("generate a valid URL", func(t *testing.T) {
		url := server.GenerateRandomURL("http", "localhost:8080", 8)
		if !server.CheckValidURL(url) {
			t.Errorf("generates an invalid url got %s", url)
		}
	})
	t.Run("generates a random URL everytime", func(t *testing.T) {
		var urlList []string
		for i := 0; i < 10; i++ {
			url := server.GenerateRandomURL("http", "localhost:8080", 8)
			for _, u := range urlList {
				if u == url {
					t.Fatalf("found duplicate urls, %s", u)
				}
			}
			urlList = append(urlList, url)
		}
	})

	t.Run("server listens on generated URL", func(t *testing.T) {
		c := cli.Newclient()
		defer c.Conn.Close()

		u := c.URL
		res, err := http.Head(u)
		if err != nil {
			t.Errorf("error making http HEAD, %s", err.Error())
		}
		if res.StatusCode == 404 {
			t.Errorf("server is not listenting on URL, %s", u)
		}
	})
}
