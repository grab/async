// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import (
	"context"
	"fmt"
)

func ExampleDoJitter() {
	t1 := InvokeInSilence(
		context.Background(), func(ctx context.Context) error {
			DoJitter(
				func() {
					fmt.Println("do something after random jitter")
				}, 1000,
			)

			return nil
		},
	)

	t2 := AddJitterT(
		NewTask(
			func(ctx context.Context) (int, error) {
				fmt.Println("return 1 after random jitter")
				return 1, nil
			},
		),
		1000,
	).Run(context.Background())

	t3 := AddJitterST(
		NewSilentTask(
			func(ctx context.Context) error {
				fmt.Println("return nil after random jitter")
				return nil
			},
		),
		1000,
	).Execute(context.Background())

	WaitAll([]SilentTask{t1, t2, t3})
}
