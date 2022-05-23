// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBatch(t *testing.T) {
	const taskCount = 10

	var wg sync.WaitGroup
	wg.Add(taskCount)

	out := make(chan int, 10)

	// Processor that multiplies items by 10 all at once
	b := NewBatcher(
		func(input []int) error {
			for _, number := range input {
				out <- number * 10
			}

			return nil
		},
	)

	for i := 0; i < taskCount; i++ {
		number := i

		ContinueInSilence(
			context.Background(), b.Append(number), func(_ context.Context, err error) error {
				defer wg.Done()

				assert.Nil(t, err)

				return nil
			},
		)
	}

	assert.Equal(t, 10, b.Size())

	b.Process()

	wg.Wait()
	close(out)

	for i := 0; i < taskCount; i++ {
		assert.Equal(t, i*10, <-out)
	}
}

func ExampleBatch() {
	var wg sync.WaitGroup
	wg.Add(2)

	b := NewBatcher(
		func(input []int) error {
			fmt.Println(input)
			return nil
		},
	)

	ContinueInSilence(
		context.Background(), b.Append(1), func(_ context.Context, err error) error {
			wg.Done()

			return nil
		},
	)

	ContinueInSilence(
		context.Background(), b.Append(2), func(_ context.Context, err error) error {
			wg.Done()

			return nil
		},
	)

	b.Process()

	wg.Wait()

	// Output:
	// [1 2]
}
