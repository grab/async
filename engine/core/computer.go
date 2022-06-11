package core

import (
	"context"

	"github.com/grab/async/async"
)

type computer interface {
	Compute(p any) async.SilentTask
}

type noisyComputer interface {
	Do(ctx context.Context, p any) (any, error)
}

type silentComputer interface {
	Do(ctx context.Context, p any) error
}

type AsyncResult struct {
	Task async.Task[any]
}

func NewAsyncResult(t async.Task[any]) AsyncResult {
	return AsyncResult{
		Task: t,
	}
}

func Take[V any](t async.Task[any]) V {
	result, _ := t.Outcome()
	return result.(V)
}

// type Computer[P any] struct {
//     computeFn func(P)
// }
//
// func (c Computer[T]) Compute(p any) async.SilentTask {
//     casted := ValidatePlan[T](p)
//
//
// }
