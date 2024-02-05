package cli_test

import (
	"testing"
	"whtester/cli"
	"whtester/server"
)

func TestWhCLI(t *testing.T) {
	t.Run("cli outputs an url", func(t *testing.T) {
		c := cli.Newclient()
		if !server.CheckValidURL(c.URL) {
			t.Fatalf("not a valid url, got %s", c.URL)
		}
	})

	t.Run("cli returns random url everytime", func(t *testing.T) {
		var urls []string
		for i := 0; i < 10; i++ {
			newURL := cli.Newclient().URL
			for _, u := range urls {
				if u == newURL {
					t.Fatalf("returned same URL, got %s", u)
				}
			}
			urls = append(urls, newURL)
		}
	})
}
