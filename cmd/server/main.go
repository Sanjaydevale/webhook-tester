package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"whtester/server"
)

type serverConfig struct {
	port   int
	domain string
}

func main() {
	conf, err := handleCmdArgs(os.Args[1:])
	if err != nil {
		log.Fatalf("invalid parameter count")
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool)

	domain := conf.domain
	port := conf.port

	clientsManager := server.NewManager()
	mux := server.NewWebHookHandler(clientsManager, domain)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	// start server
	go func() {
		srv.ListenAndServe()
	}()

	// gracefull shutdown server
	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println("Shutting Down Server")
		fmt.Println(sig)
		srv.Shutdown(context.Background())
		done <- true
	}()

	<-done
}

func handleCmdArgs(cmdArgs []string) (*serverConfig, error) {
	var conf serverConfig
	if len(cmdArgs) < 4 {
		fmt.Println("invalid arguments count")
		fmt.Println("usage main.go -d <domain> -p <port>")
		return nil, fmt.Errorf("invalid arguments count")
	}
	args := flag.NewFlagSet("args", flag.ContinueOnError)
	args.IntVar(&conf.port, "p", 8080, "port on which the server should run")
	args.StringVar(&conf.domain, "d", "localhost:8080", "domain which the server should use to generate client Urls")
	args.Parse(cmdArgs)
	return &conf, nil
}
