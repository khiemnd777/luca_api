package service

import (
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
)

type Hub struct {
	mu          sync.RWMutex
	clients     map[int][]*websocket.Conn // userID -> []*Conn
	deptClients map[int][]*websocket.Conn // deptID -> []*Conn
	Register    chan ClientConn
	Unregister  chan ClientConn
}

type ClientConn struct {
	UserID int
	DeptID int
	Conn   *websocket.Conn
}

func NewHub() *Hub {
	return &Hub{
		clients:     make(map[int][]*websocket.Conn),
		deptClients: make(map[int][]*websocket.Conn),
		Register:    make(chan ClientConn),
		Unregister:  make(chan ClientConn),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.Register:
			h.mu.Lock()
			h.clients[c.UserID] = append(h.clients[c.UserID], c.Conn)
			h.deptClients[c.DeptID] = append(h.deptClients[c.DeptID], c.Conn)
			h.mu.Unlock()
			log.Printf("✅ User %d connected", c.UserID)

		case c := <-h.Unregister:
			h.mu.Lock()
			if conns, ok := h.clients[c.UserID]; ok {
				for i, conn := range conns {
					if conn == c.Conn {
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
					if conn == c.Conn {
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
			conn.WriteMessage(websocket.TextMessage, msg)
		}
	}
}

func (h *Hub) BroadcastToUser(userID int, msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if conns, ok := h.clients[userID]; ok {
		for _, conn := range conns {
			conn.WriteMessage(websocket.TextMessage, msg)
		}
	}
}

func (h *Hub) BroadcastAll(msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, conns := range h.clients {
		for _, conn := range conns {
			conn.WriteMessage(websocket.TextMessage, msg)
		}
	}
}

func (h *Hub) BroadcastToDept(deptID int, msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if conns, ok := h.deptClients[deptID]; ok {
		for _, conn := range conns {
			conn.WriteMessage(websocket.TextMessage, msg)
		}
	}
}
