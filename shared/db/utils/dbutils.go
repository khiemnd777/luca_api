package dbutils

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/lib/pq"

	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
)

type PqStringArray []string

func (a PqStringArray) Value() (driver.Value, error) { return "{" + strings.Join(a, ",") + "}", nil }

func SortByIDs(ctx context.Context, db *sql.DB, table, orderField string, ids []int) error {
	if len(ids) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
		UPDATE %s f
		SET %s = v.ord
		FROM unnest($1::int[]) WITH ORDINALITY AS v(id, ord)
		WHERE f.id = v.id
	`,
		pq.QuoteIdentifier(table),
		pq.QuoteIdentifier(orderField),
	)

	_, err := db.ExecContext(ctx, query, pq.Array(ids))
	return err
}

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
