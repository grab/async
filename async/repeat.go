// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"time"
)

// Repeat executes the given SilentWork asynchronously on a pre-determined interval.
func Repeat(ctx context.Context, interval time.Duration, action SilentWork) SilentTask {
	safeAction := func(ctx context.Context) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic repeating task: %v \n %s", r, debug.Stack())
			}
		}()

		return action(ctx)
	}

	return InvokeInSilence(
		ctx, func(taskCtx context.Context) error {
			timer := time.NewTicker(interval)
			for {
				select {
				case <-taskCtx.Done():
					timer.Stop()
					return nil

				case <-timer.C:
					if err := safeAction(taskCtx); err != nil {
						log.Printf("error repeating task: %s", err.Error())
					}
				}
			}
		},
	)
}
