package main

import (
	"log"
	"net/http"
	"whtester/server"
)

func main() {
	mux := server.NewWebHookHandler()
	log.Fatal(http.ListenAndServe(":8080", mux))
}
