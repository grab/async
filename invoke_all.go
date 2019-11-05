// Copyright 2019 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import "context"

// InvokeAll runs the tasks with a specific max concurrency
func InvokeAll(ctx context.Context, concurrency int, tasks []Task) Task {
	if concurrency == 0 {
		return ForkJoin(ctx, tasks)
	}

	return Invoke(ctx, func(context.Context) (interface{}, error) {
		sem := make(chan struct{}, concurrency)
		for _, task := range tasks {
			sem <- struct{}{}
			task.Run(ctx).ContinueWith(ctx,
				func(interface{}, error) (interface{}, error) {
					<-sem
					return nil, nil
				})
		}
		WaitAll(tasks)
		return nil, nil
	})
}
