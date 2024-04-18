package main

import (
	"log"
	"net/http"
	"whtester/server"
)

func main() {
	clientsManager := server.NewManager()
	mux := server.NewWebHookHandler(clientsManager)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
