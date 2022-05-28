// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import (
	"context"
	"runtime"
)

func cancelRemainingTasks[T SilentTask](tasks <-chan T) {
	for {
		select {
		case t, ok := <-tasks:
			if ok {
				t.Cancel()
			}
		default:
			return
		}
	}
}

// RunWithConcurrencyLevelC runs the given tasks up to the max concurrency level.
//
// Note: When `ctx` is cancelled, we spawn a new goroutine to cancel all remaining tasks in the given channel.
// To avoid memory leak, client MUST make sure new tasks will eventually stop arriving once `ctx` is cancelled
// so that the new goroutine can return.
func RunWithConcurrencyLevelC[T SilentTask](ctx context.Context, concurrencyLevel int, tasks <-chan T) SilentTask {
	if concurrencyLevel <= 0 {
		concurrencyLevel = runtime.NumCPU()
	}

	return InvokeInSilence(
		ctx, func(taskCtx context.Context) error {
			workers := make(chan int, concurrencyLevel)
			concurrentTasks := make([]SilentTask, concurrencyLevel)

			// Generate worker IDs
			for id := 0; id < concurrencyLevel; id++ {
				workers <- id
			}

			for {
				select {
				// Context cancelled
				case <-taskCtx.Done():
					go cancelRemainingTasks(tasks)
					waitForNonNilTasks(concurrentTasks)
					return taskCtx.Err()

				// Worker available
				case workerID := <-workers:
					select {
					// Worker is waiting for job when context is cancelled
					case <-taskCtx.Done():
						go cancelRemainingTasks(tasks)
						waitForNonNilTasks(concurrentTasks)
						return taskCtx.Err()

					case t, ok := <-tasks:
						// Task channel is closed
						if !ok {
							waitForNonNilTasks(concurrentTasks)
							return nil
						}

						concurrentTasks[workerID] = t

						// Return the worker to the common pool
						ContinueInSilence(
							taskCtx, t.Execute(taskCtx), func(context.Context, error) error {
								workers <- workerID
								return nil
							},
						)
					}
				}
			}
		},
	)
}

func waitForNonNilTasks(tasks []SilentTask) {
	nonNilTasks := make([]SilentTask, 0, len(tasks))
	for _, task := range tasks {
		if task != nil {
			nonNilTasks = append(nonNilTasks, task)
		}
	}

	WaitAll(nonNilTasks)
}

// RunWithConcurrencyLevelS runs the given tasks up to the max concurrency level.
func RunWithConcurrencyLevelS[T SilentTask](ctx context.Context, concurrencyLevel int, tasks []T) SilentTask {
	if concurrencyLevel == 0 {
		concurrencyLevel = runtime.NumCPU()
	}

	return InvokeInSilence(
		ctx, func(taskCtx context.Context) error {
			sem := make(chan struct{}, concurrencyLevel)

			for i, t := range tasks {
				select {
				case <-taskCtx.Done():
					CancelAll(tasks[i:])
					return taskCtx.Err()
				case sem <- struct{}{}:
					// Return the worker to the common pool
					ContinueInSilence(
						taskCtx, t.Execute(taskCtx), func(context.Context, error) error {
							<-sem
							return nil
						},
					)
				}
			}

			WaitAll(tasks)

			return nil
		},
	)
}
