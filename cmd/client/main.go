package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"whtester/cli"
)

func main() {

	args := os.Args[1:]
	serverLink := "ws://new.whlink.sanjayj.dev/ws"
	var port int
	if len(args) == 0 {
		log.Fatalln("expected port number, usage $main <port number>")
	} else if len(args) > 1 {
		log.Fatalln("too many arguments, usage $main <port number>")
	} else {
		var err error
		port, err = strconv.Atoi(args[0])
		if err != nil {
			log.Fatalf("%v", err)
		}
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	c := cli.Newclient(serverLink)
	defer c.Conn.Close()
	fmt.Printf("link : %s", c.URL)

	// should read fields from a json file
	fields := []string{"Header", "Method", "Body"}
	go c.Stream(os.Stdout, fields, port)
	wg.Wait()
}
