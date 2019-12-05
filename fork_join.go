// Copyright 2019 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import "context"

// ForkJoin executes input task in parallel and waits for ALL outcomes before returning.
func ForkJoin(ctx context.Context, tasks []Task) Task {
	return Invoke(ctx, func(context.Context) (interface{}, error) {
		for _, task := range tasks {
			_ = task.Run(ctx)
		}
		WaitAll(tasks)
		return nil, nil
	})
}

// WaitAll waits for all tasks to finish.
func WaitAll(tasks []Task) {
	for _, task := range tasks {
		if task != nil {
			_, _ = task.Outcome()
		}
	}
}

// CancelAll cancels all specified tasks.
func CancelAll(tasks []Task) {
	for _, task := range tasks {
		task.Cancel()
	}
}
