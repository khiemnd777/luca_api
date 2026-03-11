package handler

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/order/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/middleware/rbac"
	"github.com/khiemnd777/andy_api/shared/module"
	sharedstorage "github.com/khiemnd777/andy_api/shared/storage"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type OrderDeliveryQRHandler struct {
	svc     service.OrderDeliveryQRService
	deps    *module.ModuleDeps[config.ModuleConfig]
	storage sharedstorage.Storage
}

func NewOrderDeliveryQRHandler(
	svc service.OrderDeliveryQRService,
	deps *module.ModuleDeps[config.ModuleConfig],
) *OrderDeliveryQRHandler {
	return &OrderDeliveryQRHandler{
		svc:  svc,
		deps: deps,
		storage: sharedstorage.NewLocalStorage(
			deps.Config.Storage.PhotoPath,
			"",
			"",
		),
	}
}

func (h *OrderDeliveryQRHandler) RegisterPublicRoutes(
	router fiber.Router,
	startMiddleware fiber.Handler,
	confirmMiddleware fiber.Handler,
) {
	app.RouterGet(router, "/orders/delivery/qr/:token/start", startMiddleware, h.StartSession)
	app.RouterPost(router, "/orders/delivery/confirm", confirmMiddleware, h.ConfirmDelivered)
}

func (h *OrderDeliveryQRHandler) RegisterRoutes(router fiber.Router) {
	app.RouterGet(router, "/:dept_id<int>/orders/delivery/proofs/:order_item_id<int>", h.GetDeliveryProofFile)
}

func (h *OrderDeliveryQRHandler) StartSession(c *fiber.Ctx) error {
	token := strings.TrimSpace(utils.GetParamAsString(c, "token"))
	if token == "" {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid delivery qr token")
	}

	session, err := h.svc.StartDeliveryQRSession(
		c.UserContext(),
		token,
		c.IP(),
		c.Get("User-Agent"),
	)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidDeliveryQRToken):
			return client_error.ResponseError(c, fiber.StatusUnauthorized, err, err.Error())
		case errors.Is(err, model.ErrDeliveryQRTokenAlreadyUsed):
			return client_error.ResponseError(c, fiber.StatusConflict, err, err.Error())
		case errors.Is(err, model.ErrOrderAlreadyDelivered):
			return client_error.ResponseError(c, fiber.StatusConflict, err, err.Error())
		default:
			return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
		}
	}

	c.Cookie(&fiber.Cookie{
		Name:     service.DeliveryQRSessionCookieName,
		Value:    session.SessionID,
		HTTPOnly: true,
		Secure:   true,
		SameSite: fiber.CookieSameSiteStrictMode,
		MaxAge:   int(h.sessionTTL().Seconds()),
		Path:     "/",
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message_type":       model.OrderDeliveryQRMessageTypeSessionStarted,
		"session_id":         session.SessionID,
		"order_id":           session.OrderID,
		"order_code":         session.OrderCode,
		"order_item_code":    session.OrderItemCode,
		"expires_in_seconds": int(time.Until(session.ExpiresAt).Seconds()),
		"expires_at":         session.ExpiresAt,
	})
}

func (h *OrderDeliveryQRHandler) ConfirmDelivered(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.delivery"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	sessionRaw := c.Locals("delivery_session")
	if sessionRaw == nil {
		return client_error.ResponseError(c, fiber.StatusUnauthorized, nil, "delivery session required")
	}

	session, ok := sessionRaw.(model.DeliveryQRSession)
	if !ok {
		return client_error.ResponseError(c, fiber.StatusUnauthorized, nil, "invalid delivery session")
	}

	fileHeader, err := c.FormFile("photo")
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "photo file is required")
	}

	mimeType, err := h.validateProofFile(fileHeader)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, err.Error())
	}

	relPath := service.BuildDeliveryProofStoragePath(session.OrderID, session.QRTokenID, mimeType)
	file, err := fileHeader.Open()
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "failed to read photo file")
	}
	defer file.Close()

	if _, err := h.storage.Upload(c.UserContext(), relPath, file); err != nil {
		logger.Error("delivery_confirm_failed", "order_id", session.OrderID, "qr_token_id", session.QRTokenID, "error", err.Error())
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, "failed to upload proof image")
	}

	imageFilename := filepath.Base(relPath)
	imageURL, err := h.svc.BuildDeliveryProofFileURL(c.UserContext(), session.OrderID)
	if err != nil {
		logger.Warn("delivery_proof_url_build_failed", "order_id", session.OrderID, "qr_token_id", session.QRTokenID, "error", err.Error())
	}
	logger.Info("delivery_proof_uploaded", "order_id", session.OrderID, "qr_token_id", session.QRTokenID, "image_filename", imageFilename, "mime_type", mimeType)

	err = h.svc.ConfirmDeliveredByQRSession(
		c.UserContext(),
		session.SessionID,
		imageFilename,
		fileHeader.Size,
		mimeType,
		c.IP(),
		c.Get("User-Agent"),
	)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrDeliveryQRSessionExpired):
			return client_error.ResponseError(c, fiber.StatusUnauthorized, err, err.Error())
		case errors.Is(err, model.ErrDeliveryQRSessionNotFound):
			return client_error.ResponseError(c, fiber.StatusNotFound, err, err.Error())
		case errors.Is(err, model.ErrOrderAlreadyDelivered):
			h.expireSessionCookie(c)
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"message":         "Order already delivered",
				"proof_image_url": imageURL,
			})
		case errors.Is(err, model.ErrDeliveryQRTokenAlreadyUsed):
			return client_error.ResponseError(c, fiber.StatusConflict, err, err.Error())
		case errors.Is(err, model.ErrInvalidDeliveryQRToken):
			return client_error.ResponseError(c, fiber.StatusUnauthorized, err, err.Error())
		case errors.Is(err, model.ErrDeliveryQRConfirmConcurrent):
			return client_error.ResponseError(c, fiber.StatusConflict, err, err.Error())
		default:
			return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
		}
	}

	h.expireSessionCookie(c)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":         "Order delivered successfully",
		"proof_image_url": imageURL,
	})
}

func (h *OrderDeliveryQRHandler) GetDeliveryProofFile(c *fiber.Ctx) error {
	deptID, _ := utils.GetParamAsInt(c, "dept_id")
	orderItemID, _ := utils.GetParamAsInt(c, "order_item_id")
	filePath, err := h.svc.GetDeliveryProofFilePath(c.UserContext(), deptID, orderItemID)
	if err != nil {
		if deptID <= 0 || orderItemID <= 0 {
			return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid proof image path")
		}
		return client_error.ResponseError(c, fiber.StatusNotFound, err, "proof image not found")
	}
	if _, err = os.Stat(filePath); err != nil {
		return client_error.ResponseError(c, fiber.StatusNotFound, err, "proof image not found")
	}

	return c.SendFile(filePath)
}

func (h *OrderDeliveryQRHandler) sessionTTL() time.Duration {
	minutes := h.deps.Config.DeliveryQR.SessionTTLMinutes
	if minutes <= 0 {
		return 5 * time.Minute
	}
	return time.Duration(minutes) * time.Minute
}

func (h *OrderDeliveryQRHandler) expireSessionCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     service.DeliveryQRSessionCookieName,
		Value:    "",
		HTTPOnly: true,
		Secure:   true,
		SameSite: fiber.CookieSameSiteStrictMode,
		MaxAge:   -1,
		Path:     "/",
	})
}

func (h *OrderDeliveryQRHandler) validateProofFile(fileHeader *multipart.FileHeader) (string, error) {
	maxSizeBytes := service.DeliveryProofMaxSizeBytes(h.deps.Config.DeliveryQR)
	if fileHeader.Size <= 0 || fileHeader.Size > maxSizeBytes {
		return "", fmt.Errorf("photo file exceeds the maximum allowed size")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open photo file")
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("failed to read photo file")
	}

	mimeType := http.DetectContentType(buffer[:n])
	if !service.IsAllowedDeliveryProofMimeType(mimeType) {
		return "", fmt.Errorf("unsupported photo mime type")
	}

	return mimeType, nil
}
