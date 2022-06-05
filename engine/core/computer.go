package core

import "github.com/grab/async/async"

type computer interface {
    Compute(p any) async.SilentTask
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
