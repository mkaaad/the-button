package connx

import (
	"sync"
)

var ConnPool = &connPool{
	pool: make(map[*SafeConn]bool),
	mu:   sync.RWMutex{},
}

type connPool struct {
	pool map[*SafeConn]bool
	mu   sync.RWMutex
}

func (c *connPool) Add(conn *SafeConn) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.pool[conn] = true
}

func (c *connPool) Del(conn *SafeConn) {
	c.mu.Lock()
	delete(c.pool, conn)
	c.mu.Unlock()
	conn.Close()
}

func (c *connPool) GetAllConn() (copyConns []*SafeConn) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for conn, _ := range c.pool {
		copyConns = append(copyConns, conn)
	}
	return
}
