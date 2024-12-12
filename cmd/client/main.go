package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"sync"
	"whtester/cli"
)

// ports to handle slice of ports as input
type ports []int

func (p *ports) String() string {
	return fmt.Sprintf("%v", *p)
}

func (p *ports) Set(value string) error {
	port, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("converting port string into number: %w", err)
	}

	*p = append(*p, port)
	return nil
}

var (
	defaultFields = []string{"Method", "Header", "Body"}
	serverLink    = "wss://new.whlink.sanjayj.dev/ws"
	// serverLink    = "ws://localhost:8080/ws"
	joinGroupLink = "wss://new.whlink.sanjayj.dev/wsold"
)

func main() {
	config, err := handleCmdArgs(os.Args[1:])
	if err != nil {
		log.Fatalf("handling cmd args : %s", err)
	}
	var ports []int
	for _, port := range config.ports {
		ports = append(ports, int(port))
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	var c *cli.Client
	if config.connect {
		var url string
		var key string
		fmt.Print("\nenter webhook link:")
		fmt.Scan(&url)
		fmt.Println("enter webhook password:")
		fmt.Scan(&key)
		c = cli.ConnToGroup(joinGroupLink, url, key)
	} else {
		c = cli.Newclient(serverLink)
	}
	fmt.Printf("\nlink: %s", c.URL)
	fmt.Printf("\npassword: %s", c.Key)
	defer c.Conn.Close()
	go c.Stream(os.Stdout, config.fields, ports)

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
	ports   ports
	fields  []string
	connect bool
}

func handleCmdArgs(cmdArgs []string) (*Config, error) {
	var conf Config
	args := flag.NewFlagSet("args", flag.ExitOnError)
	args.Var(&conf.ports, "p", "the port on which your webhook program is running")
	args.BoolVar(&conf.connect, "c", false, "connect to client group")
	err := args.Parse(cmdArgs)
	if err != nil {
		return nil, fmt.Errorf("parsing args : %w", err)
	}
	conf.fields, err = handleFieldArgs(args.Args())
	if err != nil {
		return nil, fmt.Errorf("handling fields : %w", err)
	}
	if len(conf.ports) == 0 {
		return nil, fmt.Errorf("expected port number")
	}

	return &conf, nil
}
