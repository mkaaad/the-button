package connx

import (
	"sync"
)

var ConnPool = &connPool{
	pool: make(map[*Client]bool),
	mu:   sync.RWMutex{},
}

type connPool struct {
	pool map[*Client]bool
	mu   sync.RWMutex
}

func (c *connPool) Add(conn *Client) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.pool[conn] = true
}

func (c *connPool) Del(conn *Client) {
	c.mu.Lock()
	delete(c.pool, conn)
	c.mu.Unlock()
	conn.Conn.Close()
}

func (c *connPool) GetAllConn() (copyConns []*Client) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for conn, _ := range c.pool {
		copyConns = append(copyConns, conn)
	}
	return
}
