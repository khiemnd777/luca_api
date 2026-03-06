package storage

import (
	"context"
	"io"
)

type Storage interface {
	Upload(ctx context.Context, path string, file io.Reader) (string, error)
}
