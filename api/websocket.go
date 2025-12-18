package api

import (
	"button/connx"
	"button/router"
	"log"
	"net/http"

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
	//userName := c.GetString("user_name")
	userName := "mkaaad"
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	client := connx.NewClient(conn)
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
		err = router.HandleMessage(message, userName, client.Send, broadcast)
		if err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte("wrong message format"))
			log.Println("Error handling message:", err)
		}
	}
}
func BroadCastMessage() {
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
