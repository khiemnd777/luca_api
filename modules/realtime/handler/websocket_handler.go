package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/khiemnd777/andy_api/modules/realtime/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/modules/realtime/realtime_model"
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
		userID := h.parseUserIDFromJWT(c)
		if userID == -1 {
			logger.Info("❌ WebSocket rejected: invalid token")
			c.Close()
			return
		}

		h.hub.Register <- service.ClientConn{UserID: userID, Conn: c}
		defer func() {
			h.hub.Unregister <- service.ClientConn{UserID: userID, Conn: c}
			c.Close()
		}()

		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				break
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
		h.hub.SendToUser(req.UserID, msg)
		return c.SendStatus(200)
	})
}

func (h *Handler) parseUserIDFromJWT(c *websocket.Conn) int {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		logger.Info("Token is empty")
		return -1
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(h.jwtSecret), nil
	})

	if err != nil {
		logger.Info(fmt.Sprintf("JWT parse error: %v", err))
		return -1
	}

	if !token.Valid {
		logger.Info(fmt.Sprintf("Token is invalid: %v", token))
		return -1
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		logger.Info(fmt.Sprintf("Map claims: %v", claims))
		return -1
	}

	if id, ok := claims["user_id"].(string); ok {
		i, err := strconv.Atoi(id)
		if err != nil {
			logger.Info(fmt.Sprintf("Int parse error: %v", err))
		}
		return i
	}

	if idFloat, ok := claims["user_id"].(float64); ok {
		if !ok {
			logger.Info(fmt.Sprintf("Float parse error: %f", idFloat))
		}
		return int(idFloat)
	}
	return -1
}
