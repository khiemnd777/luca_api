package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/order/repository"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/order"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/redis"
	"github.com/khiemnd777/andy_api/shared/utils"
)

const (
	defaultDeliveryQRSessionTTL = 5 * time.Minute
	deliveryQRRedisName         = "cache"
	deliveryQRMetaTTLBuffer     = time.Hour
	DeliveryQRSessionCookieName = "delivery_session"
	deliveryProofRootDir        = "delivery_proofs"
)

type OrderDeliveryQRService interface {
	GenerateDeliveryQRToken(ctx context.Context, orderID int) (rawToken string, err error)
	StartDeliveryQRSession(ctx context.Context, rawToken string, ip string, userAgent string) (*model.DeliveryQRSession, error)
	ConfirmDeliveredByQRSession(ctx context.Context, sessionID string, imageURL string, imageSize int64, mimeType string, ip string, userAgent string) error
}

type orderDeliveryQRService struct {
	db   *generated.Client
	repo repository.OrderDeliveryQRRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewOrderDeliveryQRService(
	db *generated.Client,
	deps *module.ModuleDeps[config.ModuleConfig],
) OrderDeliveryQRService {
	return &orderDeliveryQRService{
		db:   db,
		repo: repository.NewOrderDeliveryQRRepository(db),
		deps: deps,
	}
}

func BuildDeliveryQRStartURL(baseURL string, rawToken string) string {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	token := strings.TrimSpace(rawToken)
	if base == "" || token == "" {
		return ""
	}
	return fmt.Sprintf("%s/orders/delivery/qr/%s/start", base, token)
}

func (s *orderDeliveryQRService) GenerateDeliveryQRToken(ctx context.Context, orderID int) (string, error) {
	if orderID <= 0 {
		return "", fmt.Errorf("invalid order id")
	}

	orderEnt, err := s.db.Order.
		Query().
		Where(order.IDEQ(int64(orderID))).
		Only(ctx)
	if err != nil {
		return "", err
	}
	if strings.EqualFold(utils.DerefString(orderEnt.DeliveryStatusLatest), "delivered") {
		return "", model.ErrOrderAlreadyDelivered
	}

	rawToken := utils.GenerateRandomString(64)
	if rawToken == "" {
		return "", fmt.Errorf("failed to generate delivery qr token")
	}

	if _, err = s.repo.CreateDeliveryQRToken(ctx, nil, int64(orderID), hashDeliveryQRToken(rawToken)); err != nil {
		return "", err
	}

	logger.Info("Generated delivery QR token", "order_id", orderID)
	return rawToken, nil
}

func (s *orderDeliveryQRService) StartDeliveryQRSession(
	ctx context.Context,
	rawToken string,
	ip string,
	userAgent string,
) (*model.DeliveryQRSession, error) {
	tokenHash := hashDeliveryQRToken(rawToken)
	if tokenHash == "" {
		return nil, model.ErrInvalidDeliveryQRToken
	}

	token, err := s.repo.GetDeliveryQRTokenByHash(ctx, tokenHash)
	if err != nil {
		if generated.IsNotFound(err) {
			logger.Warn("delivery_qr_session_invalid", "ip", ip, "reason", "invalid_token")
			return nil, model.ErrInvalidDeliveryQRToken
		}
		return nil, err
	}

	orderEnt := token.Edges.Order
	if orderEnt == nil {
		logger.Warn("delivery_qr_session_invalid", "qr_token_id", token.ID, "reason", "token_without_order")
		return nil, model.ErrInvalidDeliveryQRToken
	}

	if token.Used {
		_ = s.repo.CreateDeliveryAuditLog(ctx, nil, repository.CreateOrderDeliveryAuditLogParams{
			OrderID:   token.OrderID,
			QRTokenID: intPtr(token.ID),
			Action:    model.OrderDeliveryAuditActionInvalid,
			IP:        ip,
			UserAgent: userAgent,
		})
		logger.Warn("delivery_qr_token_replay_attempt", "order_id", token.OrderID, "qr_token_id", token.ID, "ip", ip)
		return nil, model.ErrDeliveryQRTokenAlreadyUsed
	}

	if strings.EqualFold(utils.DerefString(orderEnt.DeliveryStatusLatest), "delivered") {
		_ = s.repo.CreateDeliveryAuditLog(ctx, nil, repository.CreateOrderDeliveryAuditLogParams{
			OrderID:   token.OrderID,
			QRTokenID: intPtr(token.ID),
			Action:    model.OrderDeliveryAuditActionInvalid,
			IP:        ip,
			UserAgent: userAgent,
		})
		logger.Warn("delivery_qr_session_invalid", "order_id", token.OrderID, "qr_token_id", token.ID, "reason", "order_already_delivered")
		return nil, model.ErrOrderAlreadyDelivered
	}

	now := time.Now()
	session := &model.DeliveryQRSession{
		SessionID: utils.GenerateRandomString(48),
		OrderID:   int(token.OrderID),
		QRTokenID: token.ID,
		CreatedAt: now,
		ExpiresAt: now.Add(s.sessionTTL()),
	}

	if err := s.saveSession(session); err != nil {
		return nil, err
	}

	if err := s.repo.CreateDeliveryAuditLog(ctx, nil, repository.CreateOrderDeliveryAuditLogParams{
		OrderID:   token.OrderID,
		QRTokenID: intPtr(token.ID),
		Action:    model.OrderDeliveryAuditActionScan,
		IP:        ip,
		UserAgent: userAgent,
	}); err != nil {
		return nil, err
	}

	logger.Info("delivery_qr_session_started", "order_id", token.OrderID, "qr_token_id", token.ID, "session_id", session.SessionID, "ip", ip)
	return session, nil
}

func (s *orderDeliveryQRService) ConfirmDeliveredByQRSession(
	ctx context.Context,
	sessionID string,
	imageURL string,
	imageSize int64,
	mimeType string,
	ip string,
	userAgent string,
) error {
	session, sessionErr := s.getSession(sessionID)
	if sessionErr != nil {
		if session != nil && sessionErr == model.ErrDeliveryQRSessionExpired {
			_ = s.repo.CreateDeliveryAuditLog(ctx, nil, repository.CreateOrderDeliveryAuditLogParams{
				OrderID:   int64(session.OrderID),
				QRTokenID: intPtr(session.QRTokenID),
				Action:    model.OrderDeliveryAuditActionExpired,
				IP:        ip,
				UserAgent: userAgent,
			})
			logger.Warn("delivery_confirm_failed", "session_id", sessionID, "order_id", session.OrderID, "reason", "session_expired")
			return sessionErr
		}
		logger.Warn("delivery_confirm_failed", "session_id", sessionID, "reason", "session_invalid", "error", sessionErr.Error())
		return sessionErr
	}

	now := time.Now()
	tx, err := s.db.Tx(ctx)
	if err != nil {
		logger.Error("delivery_confirm_failed", "session_id", sessionID, "error", err.Error())
		return err
	}

	if _, err := s.repo.UpsertOrderDeliveryProof(ctx, tx, repository.UpsertOrderDeliveryProofParams{
		OrderID:       int64(session.OrderID),
		QRTokenID:     session.QRTokenID,
		ImageURL:      imageURL,
		ImageSize:     imageSize,
		ImageMimeType: mimeType,
	}); err != nil {
		_ = tx.Rollback()
		logger.Error("delivery_confirm_failed", "session_id", sessionID, "order_id", session.OrderID, "error", err.Error())
		return err
	}

	updated, _, err := s.repo.UpdateOrderDelivered(ctx, tx, int64(session.OrderID), now)
	if err != nil {
		_ = tx.Rollback()
		logger.Error("delivery_confirm_failed", "session_id", sessionID, "order_id", session.OrderID, "error", err.Error())
		return err
	}

	tokenUsed, err := s.repo.MarkDeliveryQRTokenUsed(ctx, tx, session.QRTokenID, now)
	if err != nil {
		_ = tx.Rollback()
		logger.Error("delivery_confirm_failed", "session_id", sessionID, "order_id", session.OrderID, "error", err.Error())
		return err
	}
	if updated && !tokenUsed {
		_ = tx.Rollback()
		logger.Warn("delivery_qr_token_replay_attempt", "session_id", sessionID, "order_id", session.OrderID, "qr_token_id", session.QRTokenID)
		return model.ErrDeliveryQRConfirmConcurrent
	}

	if err := s.repo.CreateDeliveryAuditLog(ctx, tx, repository.CreateOrderDeliveryAuditLogParams{
		OrderID:   int64(session.OrderID),
		QRTokenID: intPtr(session.QRTokenID),
		Action:    model.OrderDeliveryAuditActionConfirm,
		IP:        ip,
		UserAgent: userAgent,
	}); err != nil {
		_ = tx.Rollback()
		logger.Error("delivery_confirm_failed", "session_id", sessionID, "order_id", session.OrderID, "error", err.Error())
		return err
	}

	if err := tx.Commit(); err != nil {
		logger.Error("delivery_confirm_failed", "session_id", sessionID, "order_id", session.OrderID, "error", err.Error())
		return err
	}
	if err := s.invalidateSession(sessionID); err != nil {
		logger.Error("delivery_confirm_failed", "session_id", sessionID, "order_id", session.OrderID, "error", err.Error())
		return err
	}

	logger.Info("delivery_confirm_success", "order_id", session.OrderID, "qr_token_id", session.QRTokenID, "session_id", sessionID, "ip", ip)
	if !updated {
		return model.ErrOrderAlreadyDelivered
	}
	return nil
}

func (s *orderDeliveryQRService) sessionTTL() time.Duration {
	minutes := s.deps.Config.DeliveryQR.SessionTTLMinutes
	if minutes <= 0 {
		return defaultDeliveryQRSessionTTL
	}
	return time.Duration(minutes) * time.Minute
}

func hashDeliveryQRToken(rawToken string) string {
	token := strings.TrimSpace(rawToken)
	if token == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func deliveryQRSessionKey(sessionID string) string {
	return fmt.Sprintf("order:delivery_session:%s", sessionID)
}

func deliveryQRSessionMetaKey(sessionID string) string {
	return fmt.Sprintf("order:delivery_session_meta:%s", sessionID)
}

func (s *orderDeliveryQRService) saveSession(session *model.DeliveryQRSession) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	if err = redis.Set(deliveryQRRedisName, deliveryQRSessionKey(session.SessionID), data, s.sessionTTL()); err != nil {
		return err
	}
	if err = redis.Set(deliveryQRRedisName, deliveryQRSessionMetaKey(session.SessionID), data, s.sessionTTL()+deliveryQRMetaTTLBuffer); err != nil {
		return err
	}

	return nil
}

func (s *orderDeliveryQRService) getSession(sessionID string) (*model.DeliveryQRSession, error) {
	return LoadDeliveryQRSession(sessionID)
}

func LoadDeliveryQRSession(sessionID string) (*model.DeliveryQRSession, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, model.ErrDeliveryQRSessionNotFound
	}

	sessionData, err := redis.Get(deliveryQRRedisName, deliveryQRSessionKey(sessionID))
	if err != nil {
		return nil, err
	}
	if sessionData == "" {
		metaData, metaErr := redis.Get(deliveryQRRedisName, deliveryQRSessionMetaKey(sessionID))
		if metaErr != nil {
			return nil, metaErr
		}
		if metaData == "" {
			return nil, model.ErrDeliveryQRSessionNotFound
		}

		session := &model.DeliveryQRSession{}
		if err = json.Unmarshal([]byte(metaData), session); err != nil {
			return nil, err
		}
		return session, model.ErrDeliveryQRSessionExpired
	}

	session := &model.DeliveryQRSession{}
	if err = json.Unmarshal([]byte(sessionData), session); err != nil {
		return nil, err
	}
	if time.Now().After(session.ExpiresAt) {
		return session, model.ErrDeliveryQRSessionExpired
	}

	return session, nil
}

func (s *orderDeliveryQRService) invalidateSession(sessionID string) error {
	if err := redis.Del(deliveryQRRedisName, deliveryQRSessionKey(sessionID)); err != nil {
		return err
	}
	if err := redis.Del(deliveryQRRedisName, deliveryQRSessionMetaKey(sessionID)); err != nil {
		return err
	}
	return nil
}

func intPtr(v int) *int {
	return &v
}

func DeliveryProofMaxSizeBytes(cfg config.DeliveryQRConfig) int64 {
	maxMB := cfg.ProofImageMaxSizeMB
	if maxMB <= 0 {
		maxMB = 5
	}
	return int64(maxMB) * 1024 * 1024
}

func BuildDeliveryProofStoragePath(orderID int, qrTokenID int, mimeType string) string {
	ext := deliveryProofExtension(mimeType)
	stableUUID := uuid.NewSHA1(uuid.NameSpaceURL, []byte(fmt.Sprintf("%d:%d", orderID, qrTokenID))).String()
	return path.Join(deliveryProofRootDir, fmt.Sprintf("%d", orderID), stableUUID+ext)
}

func BuildDeliveryProofPublicURL(baseRoute string, relPath string) string {
	cleanRelPath := path.Clean(strings.TrimLeft(relPath, "/"))
	publicPath := strings.TrimPrefix(cleanRelPath, deliveryProofRootDir+"/")
	baseRoute = strings.TrimRight(strings.TrimSpace(baseRoute), "/")
	return baseRoute + "/orders/delivery/proofs/" + publicPath
}

func deliveryProofExtension(mimeType string) string {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	default:
		return ".jpg"
	}
}

func IsAllowedDeliveryProofMimeType(mimeType string) bool {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case "image/jpeg", "image/png", "image/webp":
		return true
	default:
		return false
	}
}
