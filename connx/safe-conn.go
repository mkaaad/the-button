package connx

import (
	"sync"

	"github.com/gorilla/websocket"
)

type SafeConn struct {
	conn    *websocket.Conn
	readMu  sync.Mutex
	wirteMu sync.Mutex
}

func (sc *SafeConn) WriteMessage(messageType int, data []byte) error {
	sc.wirteMu.Lock()
	defer sc.wirteMu.Unlock()
	return sc.conn.WriteMessage(messageType, data)
}
func (sc *SafeConn) Close() error {
	return sc.conn.Close()
}
func (sc *SafeConn) ReadMessage() (int, []byte, error) {
	sc.readMu.Lock()
	defer sc.readMu.Unlock()
	return sc.conn.ReadMessage()
}
func NewSafeConn(conn *websocket.Conn) *SafeConn {
	return &SafeConn{
		conn:    conn,
		readMu:  sync.Mutex{},
		wirteMu: sync.Mutex{},
	}
}
