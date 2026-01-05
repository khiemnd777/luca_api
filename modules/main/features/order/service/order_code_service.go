package service

import (
	"context"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/features/order/repository"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
)

type OrderCodeService interface {
	ReserveOrderCode(
		ctx context.Context,
		now time.Time,
		ttl time.Duration,
	) (code string, expiresAt time.Time, err error)
	CleanupExpiredReservations(
		ctx context.Context,
		now time.Time,
	) (deleted int64, err error)
	ConfirmReservation(
		ctx context.Context,
		orderCode string,
	) error
}

type orderCodeService struct {
	db   *generated.Client
	repo repository.OrderCodeRepository
}

func NewOrderCodeService(db *generated.Client) OrderCodeService {
	return &orderCodeService{
		db:   db,
		repo: repository.NewOrderCodeRepository(db),
	}
}

func (s *orderCodeService) ReserveOrderCode(
	ctx context.Context,
	now time.Time,
	ttl time.Duration,
) (code string, expiresAt time.Time, err error) {
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return "", time.Time{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	return s.repo.ReserveOrderCode(ctx, tx, now, ttl)
}

func (s *orderCodeService) CleanupExpiredReservations(
	ctx context.Context,
	now time.Time,
) (deleted int64, err error) {
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	return s.repo.CleanupExpiredReservations(ctx, tx, now)
}

func (s *orderCodeService) ConfirmReservation(
	ctx context.Context,
	orderCode string,
) (err error) {
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	return s.repo.ConfirmReservation(ctx, tx, orderCode)
}
