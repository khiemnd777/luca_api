package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
)

type OrderCodeRepository interface {
	ReserveOrderCode(
		ctx context.Context,
		tx *generated.Tx,
		now time.Time,
		ttl time.Duration,
	) (code string, expiresAt time.Time, err error)
	ExpireReservations(
		ctx context.Context,
		tx *generated.Tx,
		now time.Time,
	) (affected int64, err error)
	CleanupExpiredReservations(
		ctx context.Context,
		tx *generated.Tx,
		now time.Time,
	) (deleted int64, err error)
	ConfirmReservation(
		ctx context.Context,
		tx *generated.Tx,
		orderCode string,
	) error
}

type orderCodeRepository struct {
	db *generated.Client
}

func NewOrderCodeRepository(db *generated.Client) OrderCodeRepository {
	return &orderCodeRepository{db: db}
}

func (r *orderCodeRepository) ReserveOrderCode(
	ctx context.Context,
	tx *generated.Tx,
	now time.Time,
	ttl time.Duration,
) (code string, expiresAt time.Time, err error) {

	period := now.Format("0106") // MMYY
	expiresAt = now.Add(ttl)

	const nextSeqSQL = `
INSERT INTO order_code_counters(period, last_seq)
VALUES ($1, 1)
ON CONFLICT (period)
DO UPDATE SET last_seq = order_code_counters.last_seq + 1
RETURNING last_seq
`

	rows, err := tx.QueryContext(ctx, nextSeqSQL, period)
	if err != nil {
		return "", time.Time{}, err
	}

	// IMPORTANT: do not defer if you will execute another statement in the same tx
	var seq int
	if !rows.Next() {
		_ = rows.Close()
		return "", time.Time{}, errors.New("failed to generate order sequence")
	}
	if err = rows.Scan(&seq); err != nil {
		_ = rows.Close()
		return "", time.Time{}, err
	}
	if err = rows.Err(); err != nil {
		_ = rows.Close()
		return "", time.Time{}, err
	}
	if err = rows.Close(); err != nil {
		return "", time.Time{}, err
	}

	code = fmt.Sprintf("%s%04d", period, seq)

	const reserveSQL = `
INSERT INTO order_code_reservations(
  order_code, period, seq, status, reserved_at, expires_at
) VALUES ($1, $2, $3, 'reserved', $4, $5)
`

	if _, err = tx.ExecContext(
		ctx,
		reserveSQL,
		code,
		period,
		seq,
		now,
		expiresAt,
	); err != nil {
		return "", time.Time{}, err
	}

	return code, expiresAt, nil
}

func (r *orderCodeRepository) ExpireReservations(
	ctx context.Context,
	tx *generated.Tx,
	now time.Time,
) (affected int64, err error) {

	const cleanupSQL = `
UPDATE order_code_reservations
SET status = 'expired'
WHERE status = 'reserved'
  AND expires_at <= $1
`

	res, err := tx.ExecContext(ctx, cleanupSQL, now)
	if err != nil {
		return 0, err
	}

	affected, err = res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return affected, nil
}

func (r *orderCodeRepository) CleanupExpiredReservations(
	ctx context.Context,
	tx *generated.Tx,
	now time.Time,
) (deleted int64, err error) {

	const cleanupSQL = `
DELETE FROM order_code_reservations
WHERE status = 'expired'
`

	res, err := tx.ExecContext(ctx, cleanupSQL, now)
	if err != nil {
		return 0, err
	}

	deleted, err = res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return deleted, nil
}

var ErrInvalidOrExpiredOrderCode = errors.New("invalid or expired order_code")

func (r *orderCodeRepository) ConfirmReservation(
	ctx context.Context,
	tx *generated.Tx,
	orderCode string,
) error {

	const confirmSQL = `
UPDATE order_code_reservations
SET status = 'used', used_at = NOW()
WHERE order_code = $1
  AND status = 'reserved'
  AND expires_at > NOW();
`

	res, err := tx.ExecContext(ctx, confirmSQL, orderCode)
	if err != nil {
		return err
	}

	affected, _ := res.RowsAffected()
	if affected == 0 {
		return ErrInvalidOrExpiredOrderCode
	}

	return nil
}
