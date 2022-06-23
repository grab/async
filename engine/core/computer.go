package core

import (
	"context"

	"github.com/grab/async/async"
)

type OutputKey struct{}

type computer interface {
	Compute(ctx context.Context, p any) (any, error)
}

type silentComputer interface {
	Compute(ctx context.Context, p any) error
}

type bridgeComputer struct {
	sc silentComputer
}

func (bc bridgeComputer) Compute(ctx context.Context, p any) (any, error) {
	return struct{}{}, bc.sc.Compute(ctx, p)
}

type AsyncResult struct {
	Task async.Task[any]
}

func newAsyncResult(t async.Task[any]) AsyncResult {
	return AsyncResult{
		Task: t,
	}
}

func Extract[V any](t async.Task[any]) V {
	result, _ := t.Outcome()
	return result.(V)
}
