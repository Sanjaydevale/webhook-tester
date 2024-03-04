package main

import (
	"fmt"
	"os"
	"sync"
	"whtester/cli"
)

func main() {
	wg := sync.WaitGroup{}
	wg.Add(1)
	c := cli.Newclient()
	defer c.Conn.Close()
	fmt.Printf("link : %s", c.URL)

	// read fields from a json file
	fields := []string{"Header", "Method", "Body"}
	go func() {
		for {
			err := c.Listen(os.Stdout, fields, ":5555")
			if err != nil {
				return
			}
		}
	}()
	wg.Wait()
}
