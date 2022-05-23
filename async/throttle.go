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
