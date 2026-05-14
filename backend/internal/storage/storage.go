package storage

import (
	"context"
	"io"
)

type Provider interface {
	Name() string
	Upload(ctx context.Context, key string, data io.Reader, contentType string, size int64) (url string, err error)
	Delete(ctx context.Context, key, mimeType string) error
	Exists(ctx context.Context, key string) (bool, error)
	Ping(ctx context.Context) error
	Usage(ctx context.Context) (any, error)
}
