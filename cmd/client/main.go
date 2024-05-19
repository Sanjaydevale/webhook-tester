package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
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

func sortFields(fieldmap map[string]struct{}) []string {
	var res []string
	for f := range fieldmap {
		res = append(res, f)
	}
	sort.Strings(res)
	return res
}

func handleFieldArgs(fields []string) ([]string, error) {
	if len(fields) == 0 {
		return defaultFields, nil
	}

	for _, field := range fields {
		_, ok := cli.AvailabeFields[field]
		if !ok {
			fmt.Println("does not contain filed ", field)
			fmt.Println("available fields :")
			for _, f := range sortFields(cli.AvailabeFields) {
				fmt.Println(f)
			}
			return nil, fmt.Errorf("checking fields")
		}
	}

	return fields, nil
}

type Config struct {
	port   int
	fields []string
}

func handleCmdArgs(cmdArgs []string) (*Config, error) {
	var conf Config
	args := flag.NewFlagSet("args", flag.ExitOnError)
	args.IntVar(&conf.port, "p", 8888, "the port on which your webhook program is running")
	err := args.Parse(cmdArgs)
	if err != nil {
		return nil, fmt.Errorf("parsing args : %w", err)
	}
	conf.fields, err = handleFieldArgs(args.Args())
	if err != nil {
		return nil, fmt.Errorf("handling fields : %w", err)
	}
	return &conf, nil
}
