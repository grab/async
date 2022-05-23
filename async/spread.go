// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import (
	"context"
	"time"
)

// Spread evenly spreads the tasks within the specified duration.
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
