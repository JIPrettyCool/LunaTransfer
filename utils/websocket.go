package utils

import (
	"LunaTransfer/models"
	"LunaTransfer/common"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	clientsMutex sync.RWMutex
	clients      = make(map[string][]*websocket.Conn)
)

type NotificationType string

const (
	NoteFileUploaded  NotificationType = "FILE_UPLOADED"
	NoteFileDownloaded NotificationType = "FILE_DOWNLOADED"
	NoteFileDeleted    NotificationType = "FILE_DELETED"
)

type Notification struct {
	Type      NotificationType `json:"type"`
	Filename  string           `json:"filename,omitempty"`
	Message   string           `json:"message"`
	Timestamp time.Time        `json:"timestamp"`
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	username, ok := common.GetUsernameFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to websocket: %v", err)
		return
	}

	clientsMutex.Lock()
	if clients[username] == nil {
		clients[username] = make([]*websocket.Conn, 0)
	}
	clients[username] = append(clients[username], conn)
	clientsMutex.Unlock()

	welcomeMsg := Notification{
		Type:      "CONNECTED",
		Message:   "Connected to LunaTransfer notifications",
		Timestamp: time.Now(),
	}

	if err := conn.WriteJSON(welcomeMsg); err != nil {
		log.Printf("Error sending welcome message: %v", err)
	}

	go handleWebSocketReader(conn, username)
}

func handleWebSocketReader(conn *websocket.Conn, username string) {
	defer func() {
		conn.Close()

		clientsMutex.Lock()
		for i, c := range clients[username] {
			if c == conn {
				clients[username] = append(clients[username][:i], clients[username][i+1:]...)
				break
			}
		}
		clientsMutex.Unlock()
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
	}
}

func NotifyUser(username string, notification models.Notification) {
	clientsMutex.RLock()
	userConns := clients[username]
	clientsMutex.RUnlock()
	if len(userConns) == 0 {
		return
	}
	notification.Timestamp = time.Now()
	data, err := json.Marshal(notification)
	if err != nil {
		log.Printf("Error marshaling notification: %v", err)
		return
	}
	for _, conn := range userConns {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("Error sending notification: %v", err)
		}
	}
}

func NotifyUpload(filename string) {
	notification := models.Notification{
		Type:     models.NoteFileUploaded,
		Filename: filename,
		Message:  fmt.Sprintf("New file uploaded: %s", filename),
	}

	clientsMutex.RLock()
	for username := range clients {
		NotifyUser(username, notification)
	}
	clientsMutex.RUnlock()
}

func NotifyFileDeleted(username, filename string) {
    notification := models.Notification{
        Type:     models.NoteFileDeleted,
        Filename: filename,
        Message:  fmt.Sprintf("File deleted: %s", filename),
    }
    
    NotifyUser(username, notification)
}
