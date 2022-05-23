// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import (
	"context"
	"reflect"
)

// ForkJoin executes given tasks in parallel and waits for ALL to complete before returning.
func ForkJoin[T SilentTask](ctx context.Context, tasks []T) SilentTask {
	return InvokeInSilence(
		ctx, func(taskCtx context.Context) error {
			for _, t := range tasks {
				t.Execute(taskCtx)
			}

			WaitAll(tasks)
			return nil
		},
	)
}

// WaitAll waits for all executed tasks to finish.
func WaitAll[T SilentTask](tasks []T) {
	for _, t := range tasks {
		if !isNil(t) {
			t.Wait()
		}
	}
}

// CancelAll cancels all given tasks.
func CancelAll[T SilentTask](tasks []T) {
	for _, t := range tasks {
		t.Cancel()
	}
}

func isNil(i interface{}) bool {
	if i == nil {
		return true
	}

	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	default:
		return false
	}
}
