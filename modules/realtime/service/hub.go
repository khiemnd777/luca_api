package service

import (
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
)

type Hub struct {
	mu         sync.RWMutex
	clients    map[int][]*websocket.Conn // userID -> []*Conn
	Register   chan ClientConn
	Unregister chan ClientConn
}

type ClientConn struct {
	UserID int
	Conn   *websocket.Conn
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[int][]*websocket.Conn),
		Register:   make(chan ClientConn),
		Unregister: make(chan ClientConn),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.Register:
			h.mu.Lock()
			h.clients[c.UserID] = append(h.clients[c.UserID], c.Conn)
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
			h.mu.Unlock()
			log.Printf("❌ User %d disconnected", c.UserID)
		}
	}
}

func (h *Hub) SendToUser(userID int, msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if conns, ok := h.clients[userID]; ok {
		for _, conn := range conns {
			conn.WriteMessage(websocket.TextMessage, msg)
		}
	}
}
