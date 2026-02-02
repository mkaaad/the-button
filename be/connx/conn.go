package connx

import (
	"github.com/gorilla/websocket"
)

type Client struct {
	Conn *websocket.Conn
	Send chan []byte
}

func (c *Client) writeLoop() {
	for msg := range c.Send {
		if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			ConnPool.Del(c)
			return
		}
	}
}
func NewClient(conn *websocket.Conn) *Client {
	c := &Client{
		Conn: conn,
		Send: make(chan []byte, 128),
	}
	go c.writeLoop()
	return c
}
