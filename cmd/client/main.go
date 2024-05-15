package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"whtester/cli"
)

func main() {

	port := flag.Int("p", 8888, "./main -p <portnumber>")
	serverLink := "ws://new.whlink.sanjayj.dev/ws"

	// check if the flag is set
	argSet := false
	flag.Parse()
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "p" {
			argSet = true
		}
	})
	if !argSet {
		log.Fatalln("expected port number, usage ./main -p <port number>")
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	c := cli.Newclient(serverLink)
	defer c.Conn.Close()
	fmt.Printf("link : %s", c.URL)

	// should read fields from a json file
	fields := []string{"Header", "Method", "Body"}
	go c.Stream(os.Stdout, fields, *port)
	wg.Wait()
}
