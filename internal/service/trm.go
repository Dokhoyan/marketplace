package service

import "context"

type TrManager interface {
	Do(ctx context.Context, fn func(context.Context) error) error
}
