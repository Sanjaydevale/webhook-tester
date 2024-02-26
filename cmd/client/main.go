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
	go func() {
		for {
			c.PrintMessage(os.Stdout)
		}
	}()
	wg.Wait()
}
