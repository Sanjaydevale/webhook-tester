package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"whtester/server"
)

func main() {

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool)

	if len(os.Args) > 3 {
		fmt.Println(os.Args)
		fmt.Println("usage main.go <domain> <port>")
		log.Fatalf("invalid parameter count")
	}
	domain := os.Args[1]
	port := os.Args[2]

	clientsManager := server.NewManager()
	mux := server.NewWebHookHandler(clientsManager, domain)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
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
