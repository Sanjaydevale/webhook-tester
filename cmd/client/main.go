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
	config, err := handleCmdArgs(os.Args[1:])
	if err != nil {
		log.Fatalf("handling cmd args : %s", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	c := cli.Newclient(serverLink)
	defer c.Conn.Close()
	fmt.Printf("link : %s", c.URL)

	// should read fields from a json file
	go c.Stream(os.Stdout, config.fields, config.port)
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

type Config struct {
	port   int
	fields []string
}

func handleCmdArgs(cmdArgs []string) (*Config, error) {
	var conf Config
	args := flag.NewFlagSet("args", flag.ContinueOnError)
	args.IntVar(&conf.port, "p", 8888, "./main -p <portnumber>")
	err := args.Parse(cmdArgs)
	if err != nil {
		return nil, fmt.Errorf("parsing args : %w", err)
	}
	conf.fields = handleFieldArgs(args.Args())
	return &conf, nil
}
