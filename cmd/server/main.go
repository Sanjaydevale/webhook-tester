package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"whtester/server"
)

func main() {

	if len(os.Args) > 3 {
		fmt.Println(os.Args)
		fmt.Println("usage main.go <domain> <port>")
		log.Fatalf("invalid parameter count")
	}
	domain := os.Args[1]
	port := os.Args[2]

	clientsManager := server.NewManager()
	mux := server.NewWebHookHandler(clientsManager, domain)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), mux))
}
