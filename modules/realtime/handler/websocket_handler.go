package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
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
}

func NewHandler(hub *service.Hub, jwtSecret string) *Handler {
	return &Handler{hub: hub, jwtSecret: jwtSecret}
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

		h.hub.Register <- service.ClientConn{UserID: userID, Conn: c}
		defer func() {
			h.hub.Unregister <- service.ClientConn{UserID: userID, Conn: c}
			c.Close()
		}()

		for {
			mt, msg, err := c.ReadMessage()
			if err != nil {
				break
			}
			if string(msg) == "ping" {
				_ = c.WriteMessage(mt, []byte("pong"))
				continue
			}
			log.Printf("📨 From user %d: %s", userID, msg)
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

		msg, _ := json.Marshal(req.Message)
		h.hub.BroadcastTo(req.UserID, msg)
		return c.SendStatus(200)
	})
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
		websocket.FormatCloseMessage(
			websocket.ClosePolicyViolation,
			reason,
		),
		time.Now().Add(time.Second),
	)
	_ = c.Close()
}
