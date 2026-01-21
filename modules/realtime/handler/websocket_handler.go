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

	// Heartbeat (message-level for proxy compatibility)
	pongWait   time.Duration // max duration allowed without receiving "pong"
	pingPeriod time.Duration // server sends "ping" every pingPeriod
	writeWait  time.Duration // write deadline for sending ping/pong/messages
}

func NewHandler(hub *service.Hub, jwtSecret string) *Handler {
	return &Handler{
		hub:       hub,
		jwtSecret: jwtSecret,

		pongWait:   60 * time.Second,
		pingPeriod: 25 * time.Second, // < pongWait
		writeWait:  5 * time.Second,
	}
}

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

		// Message-level heartbeat for proxy/gateway environments
		h.setupMessageHeartbeat(c)

		stopPing := make(chan struct{})
		defer close(stopPing)
		go h.pingLoopMessage(c, stopPing)

		for {
			mt, msg, err := c.ReadMessage()
			if err != nil {
				break
			}
			if mt != websocket.TextMessage && mt != websocket.BinaryMessage {
				continue
			}

			// We expect heartbeat as plain text: "pong" (and optionally client "ping")
			if mt == websocket.TextMessage {
				switch string(msg) {
				case "pong":
					// refresh read deadline (client is alive)
					_ = c.SetReadDeadline(time.Now().Add(h.pongWait))
					continue
				case "ping":
					// client-initiated ping; reply immediately
					_ = h.writeText(c, "pong")
					_ = c.SetReadDeadline(time.Now().Add(h.pongWait))
					continue
				}
			}

			// ignore other client messages for now
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

func (h *Handler) setupMessageHeartbeat(c *websocket.Conn) {
	// deadline = if we don't see "pong" within pongWait -> ReadMessage fails -> disconnect
	_ = c.SetReadDeadline(time.Now().Add(h.pongWait))
}

func (h *Handler) pingLoopMessage(c *websocket.Conn, stop <-chan struct{}) {
	ticker := time.NewTicker(h.pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			if err := h.writeText(c, "ping"); err != nil {
				return
			}
		}
	}
}

func (h *Handler) writeText(c *websocket.Conn, s string) error {
	_ = c.SetWriteDeadline(time.Now().Add(h.writeWait))
	return c.WriteMessage(websocket.TextMessage, []byte(s))
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

	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return -1, ErrTokenExpired
		}
	}

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
