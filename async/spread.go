// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import (
	"context"
	"time"

	"golang.org/x/time/rate"
)

// Throttle runs the given tasks at the specified rate.
func Throttle[T SilentTask](ctx context.Context, tasks []T, rateLimit int, every time.Duration) SilentTask {
	return InvokeInSilence(
		ctx, func(taskCtx context.Context) error {
			limiter := rate.NewLimiter(rate.Every(every/time.Duration(rateLimit)), 1)

			for i, t := range tasks {
				select {
				case <-taskCtx.Done():
					CancelAll(tasks[i:])
					return errCancelled
				default:
					if err := limiter.Wait(taskCtx); err == nil {
						t.Execute(taskCtx)
					}
				}
			}

			WaitAll(tasks)

			return nil
		},
	)
}

// Spread evenly spreads the work within the specified duration.
func Spread[T SilentTask](ctx context.Context, tasks []T, within time.Duration) SilentTask {
	return InvokeInSilence(
		ctx, func(taskCtx context.Context) error {
			sleep := within / time.Duration(len(tasks))

			for i, t := range tasks {
				select {
				case <-taskCtx.Done():
					CancelAll(tasks[i:])
					return errCancelled
				default:
					t.Execute(taskCtx)
					time.Sleep(sleep)
				}
			}

			WaitAll(tasks)

			return nil
		},
	)
}
