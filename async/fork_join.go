// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import (
	"context"
)

// ForkJoin executes given tasks in parallel and waits for ALL to complete before returning.
func ForkJoin[T SilentTask](ctx context.Context, tasks []T) {
	for _, t := range tasks {
		t.Execute(ctx)
	}

	WaitAll(tasks)
}

// WaitAll waits for all executed tasks to finish.
func WaitAll[T SilentTask](tasks []T) {
	for _, t := range tasks {
		t.Wait()
	}
}

// CancelAll cancels all given tasks.
func CancelAll[T SilentTask](tasks []T) {
	for _, t := range tasks {
		t.Cancel()
	}
}
