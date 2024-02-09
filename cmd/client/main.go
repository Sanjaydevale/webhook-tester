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
	fmt.Println(c.URL)
	c.PrintMessage(os.Stdout)
	wg.Wait()
}
