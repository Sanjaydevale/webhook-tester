package server_test

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"
	"whtester/cli"
	"whtester/server"

	"github.com/gorilla/websocket"
)

type clientTestFake struct {
	output [][]byte
	url    string
	domain string
	ws     *websocket.Conn
}

func (c *clientTestFake) read() {
	for {
		_, p, err := c.ws.ReadMessage()
		if err != nil {
			continue
		}
		if string(p) != "" {
			c.output = append(c.output, p)
		}
	}
}

func NewclientTestFake() *clientTestFake {
	c := &clientTestFake{}
	c.domain = "ws://localhost:8080/ws"
	ws, _, err := websocket.DefaultDialer.Dial(c.domain, nil)
	c.ws = ws
	if err != nil {
		log.Fatalf("error in clientTestFake establishing connection with whserver, %v", err.Error())
	}

	//read url
	for {
		_, p, err := c.ws.ReadMessage()
		if err != nil {
			continue
		}
		if string(p) != "" {
			c.url = string(p)
			break
		}
	}
	c.ws = ws
	return c
}

type webhookTrigger struct {
	url string
}

func (w webhookTrigger) sendPOSTRequest() {
	body := bytes.NewReader([]byte("hello world, here are the changes made in the system"))
	_, err := http.Post(w.url, "applications/json", body)
	if err != nil {
		log.Fatalf("error making post request to the server, %v", err.Error())
	}
}

func TestRandomURL(t *testing.T) {

	//assumed generated url format
	//https://Y38IpO3Ow4.example.com
	t.Run("genereates a URL with 8 character length subdomain", func(t *testing.T) {
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

		run, close := runGoFile(t, "../cmd/server/main.go", "main")
		run.Start()
		defer close()
		time.Sleep(3 * time.Second)

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

func TestForwardingMessage(t *testing.T) {
	// run server
	run, close := runGoFile(t, "../cmd/server/main.go", "main")
	run.Start()
	defer close()
	// wait for the server to start
	time.Sleep(1 * time.Second)

	t.Run("server pings the client on POST request to temp URL", func(t *testing.T) {

		// create a new client
		c := NewclientTestFake()
		go func() {
			c.read()
		}()

		// make post request to the server
		whTrigger := webhookTrigger{
			url: c.url,
		}
		whTrigger.sendPOSTRequest()

		if len(c.output) == 0 {
			t.Error("expected POST request to be forwarded by the server, but got none")
		}
	})

	t.Run("server encodes the request into binary", func(t *testing.T) {
		data := []byte("this is the body of the request")
		reqBody := bytes.NewBuffer(data)
		req, err := http.NewRequest(http.MethodPost, "http://localhost:8080", reqBody)
		if err != nil {
			t.Errorf("%v", err)
		}

		buf := server.EncodeRequest(req)
		got := server.DecodeRequest(buf)

		assertRequest(t, *got, *req)
	})
}

func assertRequest(t testing.TB, got, want http.Request) {
	t.Helper()
	fields := []string{"Method", "Proto", "ProtoMajor", "ProtoMinor", "Header", "URL", "RequestURI",
		"ContentLength", "TransferEncoding", "Host", "PostForm", "Form", "RemoteAddr"}
	assertStruct(t, fields, got, want)

	// compare bodies
	if got.Body == nil {
		got.Body = io.NopCloser(bytes.NewBuffer([]byte{}))
	}
	if want.Body == nil {
		want.Body = io.NopCloser(bytes.NewBuffer([]byte{}))
	}
	gotBody, _ := io.ReadAll(got.Body)
	wantBody, _ := io.ReadAll(want.Body)

	got.Body = io.NopCloser(bytes.NewBuffer(gotBody))
	want.Body = io.NopCloser(bytes.NewBuffer(wantBody))

	if string(gotBody) != string(wantBody) {
		t.Errorf("different Bodies, got %q, want %q", string(gotBody), string(wantBody))
	}

}

func assertStruct(t testing.TB, fields []string, structA, structB interface{}) {
	t.Helper()
	valA := reflect.ValueOf(structA)
	valB := reflect.ValueOf(structB)

	for _, field := range fields {
		if !reflect.DeepEqual(valA.FieldByName(field).Interface(), valB.FieldByName(field).Interface()) {
			t.Errorf("different %s, got %v, want %v", field, valA.FieldByName(field), valB.FieldByName(field))
		}
	}
}

func runGoFile(t testing.TB, loc string, filename string) (*exec.Cmd, func()) {
	if _, err := os.Stat(loc); err != nil {
		t.Fatalf("error running go file %s, error: %v", loc, err.Error())
	}

	buildcmd := exec.Command("go", "build", loc)
	buildcmd.Run()

	runcmd := exec.Command(fmt.Sprintf("./%s", filename))

	cleancmd := func() {
		delFile := exec.Command("rm", fmt.Sprintf("./%s", filename))
		delFile.Run()
		runcmd.Process.Kill()
	}
	return runcmd, cleancmd
}
