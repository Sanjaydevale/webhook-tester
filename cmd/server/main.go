package main

import (
	"log"
	"net/http"
	"whtester/server"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Fatalf("error establishing websocket connection")
		}
		u := []byte(server.GenerateRandomURL("http", "localhost:8080", 8))
		ws.WriteMessage(websocket.TextMessage, u)
	})
	http.ListenAndServe(":8080", mux)
}
