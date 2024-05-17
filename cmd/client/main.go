package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"whtester/cli"
)

var (
	defaultFields = []string{"Method", "Header", "Body"}
	serverLink    = "ws://new.whlink.sanjayj.dev/ws"
)

func main() {

	port := flag.Int("p", 8888, "./main -p <portnumber>")
	fieldCmd := flag.NewFlagSet("field", flag.ExitOnError)

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

	// parse the field values
	fieldCmd.Parse(os.Args[3:])
	fields := fieldCmd.Args()
	fields = handleFieldArgs(fields)

	wg := sync.WaitGroup{}
	wg.Add(1)
	c := cli.Newclient(serverLink)
	defer c.Conn.Close()
	fmt.Printf("link : %s", c.URL)

	// should read fields from a json file
	go c.Stream(os.Stdout, fields, *port)
	wg.Wait()
}

func handleFieldArgs(fields []string) []string {
	if len(fields) == 0 {
		return defaultFields
	}

	for _, field := range fields {
		_, ok := cli.AvailabeFields[field]
		if !ok {
			fmt.Println("does not contain filed ", field)
			fmt.Println("available fields : ")
			for f := range cli.AvailabeFields {
				fmt.Println(f)
			}
			os.Exit(0)
		}
	}

	return fields
}
