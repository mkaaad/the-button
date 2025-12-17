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
	broadcast = make(chan []byte, 100) // Broadcast channel
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
	userID := c.GetInt64("user_id")
	unsafeConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	conn := connx.NewSafeConn(unsafeConn)
	// Register new client
	connx.ConnPool.Add(conn)
	// Ensure connection is closed on function exit
	defer func() {
		connx.ConnPool.Del(conn)
	}()

	for {
		// Read message from browser
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			break
		}
		if messageType != websocket.TextMessage {
			continue // Ignore non-text messages
		}
		// Handle the message
		err = router.HandleMessage(conn, message, userID, broadcast)
		if err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte("wrong message format"))
			log.Println("Error handling message:", err)
		}
	}
}
func BroadCastMessage() {
	for {
		message := <-broadcast
		conns := connx.ConnPool.GetAllConn()
		for _, conn := range conns {
			go func(c *connx.SafeConn) {
				err := c.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					go connx.ConnPool.Del(c)
				}
			}(conn)
		}
	}
}
