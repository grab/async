package core

import "context"

type plan interface {
	Execute(ctx context.Context) error
	IsSequential() bool
}
