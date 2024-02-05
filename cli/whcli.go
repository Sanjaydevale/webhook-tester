package cli

import (
	"log"

	"github.com/gorilla/websocket"
)

func NewURL() string {
	ws, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		log.Fatalf("error establishing websocket connection: %v", err.Error())
	}
	defer ws.Close()

	for {
		_, data, err := ws.ReadMessage()
		if err != nil {
			continue
		}
		if len(data) != 0 {
			return string(data)
		}
	}
}
