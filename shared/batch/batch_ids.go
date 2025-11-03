package batchUtil

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/khiemnd777/andy_api/shared/app"
)

const maxConcurrentBatch = 20

func BatchGetByIDs[T any](
	ctx context.Context,
	ids []int,
	buildGetFunc func(id int) func() (*T, error),
) ([]*T, error) {
	var (
		results []*T
		mu      sync.Mutex
		wg      sync.WaitGroup
	)

	errChan := make(chan error, len(ids))
	sem := make(chan struct{}, maxConcurrentBatch)

	for _, id := range ids {
		wg.Add(1)
		sem <- struct{}{}

		go func(id int) {
			defer func() {
				<-sem
				wg.Done()
			}()

			getFunc := buildGetFunc(id)

			entity, err := getFunc()
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "not found") {
					return
				}

				errChan <- err
				return
			}

			if entity != nil {
				mu.Lock()
				results = append(results, entity)
				mu.Unlock()
			}
		}(id)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		return nil, err
	}

	return results, nil
}

func BatchApiGetByIDs[T any](ctx context.Context, module string, access string, ids []int) ([]*T, error) {
	body := map[string]any{
		"ids": ids,
	}

	var result []*T

	if err := app.GetHttpClient().CallPost(ctx, module, fmt.Sprintf("/api/%s/batch-get", module), access, "", body, &result); err != nil {
		return nil, err
	}

	return result, nil
}
