package state

import "context"

type Cache interface {
	GetState(ctx context.Context, key string) (string, error)
	SetState(ctx context.Context, key string, value string) error
	DeleteState(ctx context.Context, key string) error
}
