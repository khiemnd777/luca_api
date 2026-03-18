package service

import (
	"log"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
)

type Hub struct {
	mu          sync.RWMutex
	clients     map[int][]*ClientConn // userID -> []*Conn
	deptClients map[int][]*ClientConn // deptID -> []*Conn
	Register    chan *ClientConn
	Unregister  chan *ClientConn
}

type ClientConn struct {
	UserID  int
	DeptID  int
	Conn    *websocket.Conn
	writeMu sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients:     make(map[int][]*ClientConn),
		deptClients: make(map[int][]*ClientConn),
		Register:    make(chan *ClientConn),
		Unregister:  make(chan *ClientConn),
	}
}

func (c *ClientConn) WriteMessage(messageType int, msg []byte, deadline time.Duration) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if deadline > 0 {
		_ = c.Conn.SetWriteDeadline(time.Now().Add(deadline))
	}
	return c.Conn.WriteMessage(messageType, msg)
}

func (c *ClientConn) WriteControl(messageType int, data []byte, deadline time.Duration) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	writeDeadline := time.Now()
	if deadline > 0 {
		writeDeadline = writeDeadline.Add(deadline)
	}
	return c.Conn.WriteControl(messageType, data, writeDeadline)
}

func (c *ClientConn) Close() error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.Conn.Close()
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.Register:
			h.mu.Lock()
			h.clients[c.UserID] = append(h.clients[c.UserID], c)
			h.deptClients[c.DeptID] = append(h.deptClients[c.DeptID], c)
			h.mu.Unlock()
			log.Printf("✅ User %d connected", c.UserID)

		case c := <-h.Unregister:
			h.mu.Lock()
			if conns, ok := h.clients[c.UserID]; ok {
				for i, conn := range conns {
					if conn == c {
						h.clients[c.UserID] = append(conns[:i], conns[i+1:]...)
						break
					}
				}
				if len(h.clients[c.UserID]) == 0 {
					delete(h.clients, c.UserID)
				}
			}
			if conns, ok := h.deptClients[c.DeptID]; ok {
				for i, conn := range conns {
					if conn == c {
						h.deptClients[c.DeptID] = append(conns[:i], conns[i+1:]...)
						break
					}
				}
				if len(h.deptClients[c.DeptID]) == 0 {
					delete(h.deptClients, c.DeptID)
				}
			}
			h.mu.Unlock()
			log.Printf("❌ User %d disconnected", c.UserID)
		}
	}
}

// [obsoleted] use BroadcastTo
func (h *Hub) SendToUser(userID int, msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if conns, ok := h.clients[userID]; ok {
		for _, conn := range conns {
			_ = conn.WriteMessage(websocket.TextMessage, msg, 5*time.Second)
		}
	}
}

func (h *Hub) BroadcastToUser(userID int, msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if conns, ok := h.clients[userID]; ok {
		for _, conn := range conns {
			_ = conn.WriteMessage(websocket.TextMessage, msg, 5*time.Second)
		}
	}
}

func (h *Hub) BroadcastAll(msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, conns := range h.clients {
		for _, conn := range conns {
			_ = conn.WriteMessage(websocket.TextMessage, msg, 5*time.Second)
		}
	}
}

func (h *Hub) BroadcastToDept(deptID int, msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if conns, ok := h.deptClients[deptID]; ok {
		for _, conn := range conns {
			_ = conn.WriteMessage(websocket.TextMessage, msg, 5*time.Second)
		}
	}
}
