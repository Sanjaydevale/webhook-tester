package cli_test

import (
	"testing"
	"whtester/cli"
	"whtester/server"
)

func TestWhCLI(t *testing.T) {
	t.Run("cli outputs an url", func(t *testing.T) {
		url := cli.NewURL()
		if !server.CheckValidURL(url) {
			t.Fatalf("not a valid url, got %s", url)
		}
	})

	t.Run("cli returns random url everytime", func(t *testing.T) {
		var urls []string
		for i := 0; i < 10; i++ {
			newURL := cli.NewURL()
			for _, u := range urls {
				if u == newURL {
					t.Fatalf("returned same URL, got %s", u)
				}
			}
			urls = append(urls, newURL)
		}
	})
}
