package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/khiemnd777/andy_api/modules/realtime/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/modules/realtime/realtime_model"
)

var (
	ErrTokenExpired = errors.New("token_expired")
	ErrTokenInvalid = errors.New("token_invalid")
)

type Handler struct {
	hub       *service.Hub
	jwtSecret string

	// heartbeat settings
	pongWait   time.Duration
	pingPeriod time.Duration
	writeWait  time.Duration
}

func NewHandler(hub *service.Hub, jwtSecret string) *Handler {
	// Recommended defaults:
	// - pingPeriod < pongWait
	// - writeWait small to avoid blocking forever on a bad client
	return &Handler{
		hub:       hub,
		jwtSecret: jwtSecret,

		pongWait:   60 * time.Second,
		pingPeriod: 30 * time.Second,
		writeWait:  5 * time.Second,
	}
}

// Optional if you want to tune from config
func (h *Handler) WithHeartbeat(pongWait, pingPeriod, writeWait time.Duration) *Handler {
	if pongWait > 0 {
		h.pongWait = pongWait
	}
	if pingPeriod > 0 {
		h.pingPeriod = pingPeriod
	}
	if writeWait > 0 {
		h.writeWait = writeWait
	}
	return h
}

func (h *Handler) RegisterRoutes(router fiber.Router) {
	// External WS endpoint
	app.RouterGet(router, "/", websocket.New(func(c *websocket.Conn) {
		userID, err := h.parseUserIDFromJWT(c)
		if err != nil {
			if errors.Is(err, ErrTokenExpired) {
				h.closeWithReason(c, "token_expired")
			} else {
				h.closeWithReason(c, "token_invalid")
			}
			return
		}

		client := service.ClientConn{UserID: userID, Conn: c}

		h.hub.Register <- client
		defer func() {
			h.hub.Unregister <- client
			_ = c.Close()
		}()

		// ==== WebSocket lifecycle hardening ====
		h.setupHeartbeat(c)

		// Ping loop: server-driven
		stopPing := make(chan struct{})
		defer close(stopPing)
		go h.pingLoop(c, stopPing)

		// Read loop
		for {
			mt, msg, err := c.ReadMessage()
			if err != nil {
				// normal: client disconnect / network error / read timeout
				break
			}

			// If you only accept text messages from client, enforce it:
			// - You can drop binary frames if not needed.
			if mt != websocket.TextMessage && mt != websocket.BinaryMessage {
				continue
			}

			// TODO: If later you want client -> server actions, parse msg here.
			// For now we only log or ignore to avoid unnecessary work.
			_ = msg
		}
	}))
}

func (h *Handler) RegisterInternalRoutes(router fiber.Router) {
	app.RouterPost(router, "/internal/send", func(c *fiber.Ctx) error {
		var req struct {
			UserID  int                             `json:"user_id"`
			Message realtime_model.RealtimeEnvelope `json:"message"`
		}

		if err := c.BodyParser(&req); err != nil {
			logger.Debug(fmt.Sprintf("ERROR: %v", err))
			return fiber.ErrBadRequest
		}

		msg, err := json.Marshal(req.Message)
		if err != nil {
			logger.Debug(fmt.Sprintf("ERROR: %v", err))
			return fiber.ErrBadRequest
		}

		h.hub.BroadcastTo(req.UserID, msg)
		return c.SendStatus(200)
	})
}

// setupHeartbeat configures:
// - read deadline
// - pong handler that refreshes the deadline
func (h *Handler) setupHeartbeat(c *websocket.Conn) {
	// Limit payload to protect server (tune if needed)
	// c.SetReadLimit(64 * 1024)

	_ = c.SetReadDeadline(time.Now().Add(h.pongWait))
	c.SetPongHandler(func(string) error {
		// Each pong means client is alive
		return c.SetReadDeadline(time.Now().Add(h.pongWait))
	})

	// If you want to observe pings from client (not required):
	// c.SetPingHandler(func(appData string) error {
	//   _ = c.SetReadDeadline(time.Now().Add(h.pongWait))
	//   // Reply pong immediately
	//   _ = c.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(h.writeWait))
	//   return nil
	// })
}

func (h *Handler) pingLoop(c *websocket.Conn, stop <-chan struct{}) {
	ticker := time.NewTicker(h.pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			// Control frames should not block long
			deadline := time.Now().Add(h.writeWait)
			if err := c.WriteControl(websocket.PingMessage, []byte{}, deadline); err != nil {
				// any ping error indicates the connection is likely broken
				return
			}
		}
	}
}

func (h *Handler) parseUserIDFromJWT(c *websocket.Conn) (int, error) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		return -1, ErrTokenInvalid
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(h.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return -1, ErrTokenInvalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return -1, ErrTokenInvalid
	}

	// check exp
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return -1, ErrTokenExpired
		}
	}

	// extract user_id
	switch v := claims["user_id"].(type) {
	case string:
		id, err := strconv.Atoi(v)
		if err != nil {
			return -1, ErrTokenInvalid
		}
		return id, nil
	case float64:
		return int(v), nil
	}

	return -1, ErrTokenInvalid
}

func (h *Handler) closeWithReason(c *websocket.Conn, reason string) {
	_ = c.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.ClosePolicyViolation, reason),
		time.Now().Add(time.Second),
	)
	_ = c.Close()
}
