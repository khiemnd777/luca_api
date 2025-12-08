package dbutils

import (
	"context"
	"database/sql/driver"
	"strings"

	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
)

type PqStringArray []string

func (a PqStringArray) Value() (driver.Value, error) { return "{" + strings.Join(a, ",") + "}", nil }

func WithTx[D any](ctx context.Context, db *generated.Client, fn func(tx *generated.Tx) (D, error)) (D, error) {
	var (
		tx   *generated.Tx
		err  error
		zero D
	)

	tx, err = db.Tx(ctx)
	if err != nil {
		return zero, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		if txErr := tx.Commit(); txErr != nil {
			err = txErr
		}
	}()

	var out D
	out, err = fn(tx)
	return out, err
}
