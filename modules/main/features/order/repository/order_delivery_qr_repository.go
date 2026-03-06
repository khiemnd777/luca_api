package repository

import (
	"context"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/order"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/orderdeliveryproof"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/orderdeliveryqrtoken"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/orderitem"
)

type CreateOrderDeliveryAuditLogParams struct {
	OrderID   int64
	QRTokenID *int
	Action    model.OrderDeliveryAuditAction
	IP        string
	UserAgent string
}

type UpsertOrderDeliveryProofParams struct {
	OrderID       int64
	QRTokenID     int
	ImageURL      string
	ImageSize     int64
	ImageMimeType string
}

type OrderDeliveryQRRepository interface {
	CreateDeliveryQRToken(ctx context.Context, tx *generated.Tx, orderID int64, tokenHash string) (*generated.OrderDeliveryQRToken, error)
	GetDeliveryQRTokenByHash(ctx context.Context, tokenHash string) (*generated.OrderDeliveryQRToken, error)
	GetOrderDeliveryProofByQRTokenID(ctx context.Context, qrTokenID int) (*generated.OrderDeliveryProof, error)
	UpsertOrderDeliveryProof(ctx context.Context, tx *generated.Tx, params UpsertOrderDeliveryProofParams) (*generated.OrderDeliveryProof, error)
	MarkDeliveryQRTokenUsed(ctx context.Context, tx *generated.Tx, qrTokenID int, usedAt time.Time) (bool, error)
	UpdateOrderDelivered(ctx context.Context, tx *generated.Tx, orderID int64, deliveredAt time.Time) (bool, int64, error)
	CreateDeliveryAuditLog(ctx context.Context, tx *generated.Tx, params CreateOrderDeliveryAuditLogParams) error
}

type orderDeliveryQRRepository struct {
	db *generated.Client
}

func NewOrderDeliveryQRRepository(db *generated.Client) OrderDeliveryQRRepository {
	return &orderDeliveryQRRepository{db: db}
}

func (r *orderDeliveryQRRepository) CreateDeliveryQRToken(
	ctx context.Context,
	tx *generated.Tx,
	orderID int64,
	tokenHash string,
) (*generated.OrderDeliveryQRToken, error) {
	client := r.db.OrderDeliveryQRToken
	if tx != nil {
		client = tx.OrderDeliveryQRToken
	}

	return client.
		Create().
		SetOrderID(orderID).
		SetTokenHash(tokenHash).
		Save(ctx)
}

func (r *orderDeliveryQRRepository) GetDeliveryQRTokenByHash(
	ctx context.Context,
	tokenHash string,
) (*generated.OrderDeliveryQRToken, error) {
	return r.db.OrderDeliveryQRToken.
		Query().
		Where(orderdeliveryqrtoken.TokenHashEQ(tokenHash)).
		WithOrder().
		Only(ctx)
}

func (r *orderDeliveryQRRepository) GetOrderDeliveryProofByQRTokenID(
	ctx context.Context,
	qrTokenID int,
) (*generated.OrderDeliveryProof, error) {
	return r.db.OrderDeliveryProof.
		Query().
		Where(orderdeliveryproof.QrTokenIDEQ(qrTokenID)).
		Only(ctx)
}

func (r *orderDeliveryQRRepository) UpsertOrderDeliveryProof(
	ctx context.Context,
	tx *generated.Tx,
	params UpsertOrderDeliveryProofParams,
) (*generated.OrderDeliveryProof, error) {
	client := r.db.OrderDeliveryProof
	if tx != nil {
		client = tx.OrderDeliveryProof
	}

	existing, err := client.
		Query().
		Where(orderdeliveryproof.QrTokenIDEQ(params.QRTokenID)).
		Only(ctx)
	if err != nil && !generated.IsNotFound(err) {
		return nil, err
	}

	if existing != nil {
		return client.
			UpdateOneID(existing.ID).
			SetImageURL(params.ImageURL).
			SetImageSize(params.ImageSize).
			SetImageMimeType(params.ImageMimeType).
			Save(ctx)
	}

	return client.
		Create().
		SetOrderID(params.OrderID).
		SetQrTokenID(params.QRTokenID).
		SetImageURL(params.ImageURL).
		SetImageSize(params.ImageSize).
		SetImageMimeType(params.ImageMimeType).
		Save(ctx)
}

func (r *orderDeliveryQRRepository) MarkDeliveryQRTokenUsed(
	ctx context.Context,
	tx *generated.Tx,
	qrTokenID int,
	usedAt time.Time,
) (bool, error) {
	client := r.db.OrderDeliveryQRToken
	if tx != nil {
		client = tx.OrderDeliveryQRToken
	}

	affected, err := client.
		Update().
		Where(
			orderdeliveryqrtoken.IDEQ(qrTokenID),
			orderdeliveryqrtoken.Used(false),
		).
		SetUsed(true).
		SetUsedAt(usedAt).
		Save(ctx)
	if err != nil {
		return false, err
	}

	return affected > 0, nil
}

func (r *orderDeliveryQRRepository) UpdateOrderDelivered(
	ctx context.Context,
	tx *generated.Tx,
	orderID int64,
	deliveredAt time.Time,
) (bool, int64, error) {
	orderItemClient := r.db.OrderItem
	orderClient := r.db.Order
	if tx != nil {
		orderItemClient = tx.OrderItem
		orderClient = tx.Order
	}

	latestItem, err := orderItemClient.
		Query().
		Where(
			orderitem.OrderID(orderID),
			orderitem.DeletedAtIsNil(),
		).
		Order(generated.Desc(orderitem.FieldCreatedAt)).
		First(ctx)
	if err != nil {
		return false, 0, err
	}

	updatedItems, err := orderItemClient.
		Update().
		Where(
			orderitem.IDEQ(latestItem.ID),
			orderitem.DeletedAtIsNil(),
			orderitem.Or(
				orderitem.DeliveryStatusIsNil(),
				orderitem.DeliveryStatusNEQ("delivered"),
			),
		).
		SetDeliveryStatus("delivered").
		SetDeliveredAt(deliveredAt).
		Save(ctx)
	if err != nil {
		return false, 0, err
	}
	if updatedItems == 0 {
		return false, latestItem.ID, nil
	}

	if _, err = orderClient.
		Update().
		Where(
			order.IDEQ(orderID),
			order.Or(
				order.DeliveryStatusLatestIsNil(),
				order.DeliveryStatusLatestNEQ("delivered"),
			),
		).
		SetDeliveryStatusLatest("delivered").
		Save(ctx); err != nil {
		return false, 0, err
	}

	return true, latestItem.ID, nil
}

func (r *orderDeliveryQRRepository) CreateDeliveryAuditLog(
	ctx context.Context,
	tx *generated.Tx,
	params CreateOrderDeliveryAuditLogParams,
) error {
	client := r.db.OrderDeliveryAuditLog
	if tx != nil {
		client = tx.OrderDeliveryAuditLog
	}

	q := client.
		Create().
		SetOrderID(params.OrderID).
		SetAction(string(params.Action))

	if params.QRTokenID != nil {
		q.SetQrTokenID(*params.QRTokenID)
	}
	if params.IP != "" {
		q.SetIP(params.IP)
	}
	if params.UserAgent != "" {
		q.SetUserAgent(params.UserAgent)
	}

	_, err := q.Save(ctx)
	return err
}
