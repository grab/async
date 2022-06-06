package core

import (
	"context"

	"github.com/grab/async/async"
)

type computer interface {
	Compute(p any) async.SilentTask
}

type syncComputer interface {
	Compute(ctx context.Context, p any) error
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
