// Copyright 2019 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import (
	"context"
	"runtime"
)

// Consume runs the tasks with a specific max concurrency
func Consume(ctx context.Context, concurrency int, tasks chan Task) Task {
	if concurrency <= 0 {
		concurrency = runtime.NumCPU()
	}

	return Invoke(ctx, func(context.Context) (interface{}, error) {
		workers := make(chan int, concurrency)
		concurrentTasks := make([]Task, concurrency)
		// generate worker IDs
		for id := 0; id < concurrency; id++ {
			workers <- id
		}

		for {
			select {
			// context cancelled
			case <-ctx.Done():
				WaitAll(concurrentTasks)
				return nil, ctx.Err()

				// worker available
			case workerID := <-workers:
				select {
				// worker is waiting for job when context is cancelled
				case <-ctx.Done():
					WaitAll(concurrentTasks)
					return nil, ctx.Err()

				case t, ok := <-tasks:
					// if task channel is closed
					if !ok {
						WaitAll(concurrentTasks)
						return nil, nil
					}
					concurrentTasks[workerID] = t
					t.Run(ctx).ContinueWith(ctx,
						func(interface{}, error) (interface{}, error) {
							workers <- workerID
							return nil, nil
						})
				}
			}
		}
	})
}
