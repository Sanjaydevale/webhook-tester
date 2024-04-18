package main

import (
	"log"
	"net/http"
	"whtester/server"
)

func main() {
	clientsManager := &server.Manager{}
	mux := server.NewWebHookHandler(clientsManager)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
