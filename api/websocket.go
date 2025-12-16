package api

import (
	"button/router"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	clients   = make(map[*websocket.Conn]bool) // Connected clients
	clientMu  sync.RWMutex                     // Mutex for clients map
	broadcast = make(chan []byte, 100)         // Broadcast channel
)

// Upgrader is used to upgrade HTTP connections to WebSocket connections.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // Allow all origins for simplicity
}

// WebSocketHandler handles WebSocket upgrade requests.
func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	userID, ok := authenticateToken(token)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Upgrade initial GET request to a websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}
	// Register new client
	clientMu.Lock()
	clients[conn] = true
	clientMu.Unlock()

	// Ensure connection is closed on function exit
	defer func() {
		clientMu.Lock()
		delete(clients, conn)
		clientMu.Unlock()
		conn.Close()
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
func authenticateToken(token string) (int64, bool) {
	// Placeholder for token authentication logic
	// In a real application, verify the token against your authentication system
	return 0, true
}
func BroadCastMessage() {
	// 持续监听广播通道
	for {
		// 从广播通道取出消息（阻塞等待）
		message := <-broadcast

		// 第一步：加读锁，快速复制客户端集合到本地
		clientMu.RLock()
		// 复制当前所有客户端连接到切片，避免遍历原map时的并发问题
		clientList := make([]*websocket.Conn, 0, len(clients))
		for conn := range clients {
			clientList = append(clientList, conn)
		}
		// 立即释放读锁，减少锁持有时间（核心优化）
		clientMu.RUnlock()

		// 第二步：遍历本地切片，异步推送消息给每个客户端
		for _, conn := range clientList {
			// 关键：将conn作为参数传入协程，避免循环变量捕获陷阱
			go func(c *websocket.Conn) {
				// 向客户端推送消息
				err := c.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					// 推送失败：加写锁移除客户端并关闭连接
					clientMu.Lock()
					// 先删除再关闭，避免重复操作
					delete(clients, c)
					clientMu.Unlock()
					// 关闭连接释放资源
					c.Close()
				}
			}(conn)
		}
	}
}
