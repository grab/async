// Copyright 2019 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import (
	"context"
	"time"

	"golang.org/x/time/rate"
)

// Throttle runs the tasks with a specified rate limiter.
func Throttle(ctx context.Context, tasks []Task, rateLimit int, every time.Duration) Task {
	return Invoke(ctx, func(context.Context) (interface{}, error) {
		limiter := rate.NewLimiter(rate.Every(every/time.Duration(rateLimit)), 1)
		for i, task := range tasks {
			select {
			case <-ctx.Done():
				CancelAll(tasks[i:])
				return nil, errCancelled
			default:
				if err := limiter.Wait(ctx); err == nil {
					task.Run(ctx)
				}
			}
		}

		WaitAll(tasks)
		return nil, nil
	})
}

// Spread evenly spreads the work within the specified duration.
func Spread(ctx context.Context, within time.Duration, tasks []Task) Task {
	return Invoke(ctx, func(context.Context) (interface{}, error) {
		sleep := within / time.Duration(len(tasks))
		for _, task := range tasks {
			select {
			case <-ctx.Done():
				return nil, errCancelled
			default:
				task.Run(ctx)
				time.Sleep(sleep)
			}
		}

		WaitAll(tasks)
		return nil, nil
	})
}
