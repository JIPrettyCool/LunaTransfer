package utils

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
)

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan string)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "WebSocket upgrade failed", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	clients[conn] = true
	for msg := range broadcast {
		for client := range clients {
			client.WriteMessage(websocket.TextMessage, []byte(msg))
		}
	}
}

func NotifyUpload(filename string) {
	msg := fmt.Sprintf("New file uploaded: %s", filename)
	broadcast <- msg
}
