package api

import (
	"button/config"
	"button/connx"
	"button/service"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	broadcast = make(chan []byte, 2000) // Broadcast channel
)

// Upgrader is used to upgrade HTTP connections to WebSocket connections.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // Allow all origins for simplicity
}

// WebSocketHandler handles WebSocket upgrade requests.
func WebSocketHandler(c *gin.Context) {
	// Upgrade initial GET request to a websocket
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"info": "未登录",
		})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	client := connx.NewClient(conn, sessionID)
	// Register new client
	connx.ConnPool.Add(client)
	// Ensure connection is closed on function exit
	defer func() {
		connx.ConnPool.Del(client)
	}()

	for {
		// Read message from browser
		messageType, message, err := client.Conn.ReadMessage()
		if err != nil {
			break
		}
		if messageType != websocket.TextMessage {
			continue // Ignore non-text messages
		}
		// Handle the message
		err = handleMessage(message, client.SessionID, client.Send, broadcast)
		if err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte("wrong message format"))
			log.Println("Error handling message:", err)
		}
	}
}
func BroadCastMessage() {
	go checkTimeCron()
	for msg := range broadcast {
		conns := connx.ConnPool.GetAllConn()
		for _, c := range conns {
			select {
			case c.Send <- msg:
			default:
				go connx.ConnPool.Del(c)
			}
		}
	}
}
func checkTimeCron() {
	t := time.NewTicker(200 * time.Millisecond)
	for range t.C {
		if time.Now().After(config.StartTime) {
			service.GetTime(broadcast)
			break
		}
	}
}
