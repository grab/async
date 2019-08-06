// Copyright (c) 2012-2019 Grabtaxi Holdings PTE LTD (GRAB), All Rights Reserved. NOTICE: All information contained herein
// is, and remains the property of GRAB. The intellectual and technical concepts contained herein are confidential, proprietary
// and controlled by GRAB and may be covered by patents, patents in process, and are protected by trade secret or copyright law.
//
// You are strictly forbidden to copy, download, store (in any medium), transmit, disseminate, adapt or change this material
// in any way unless prior written permission is obtained from GRAB. Access to the source code contained herein is hereby
// forbidden to anyone except current GRAB employees or contractors with binding Confidentiality and Non-disclosure agreements
// explicitly covering such access.
//
// The copyright notice above does not evidence any actual or intended publication or disclosure of this source code,
// which includes information that is confidential and/or proprietary, and is a trade secret, of GRAB.
//
// ANY REPRODUCTION, MODIFICATION, DISTRIBUTION, PUBLIC PERFORMANCE, OR PUBLIC DISPLAY OF OR THROUGH USE OF THIS SOURCE
// CODE WITHOUT THE EXPRESS WRITTEN CONSENT OF GRAB IS STRICTLY PROHIBITED, AND IN VIOLATION OF APPLICABLE LAWS AND
// INTERNATIONAL TREATIES. THE RECEIPT OR POSSESSION OF THIS SOURCE CODE AND/OR RELATED INFORMATION DOES NOT CONVEY
// OR IMPLY ANY RIGHTS TO REPRODUCE, DISCLOSE OR DISTRIBUTE ITS CONTENTS, OR TO MANUFACTURE, USE, OR SELL ANYTHING
// THAT IT MAY DESCRIBE, IN WHOLE OR IN PART.

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
