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

	// Reducer that multiplies items by 10 at once
	b := NewBatch(
		context.Background(), func(input []int) []int {
			result := make([]int, len(input))
			for i, number := range input {
				result[i] = number * 10
			}

			return result
		},
	)

	for i := 0; i < taskCount; i++ {
		number := i

		ContinueWithNoResult(
			context.Background(), b.Append(number), func(_ context.Context, result int, err error) error {
				defer wg.Done()

				assert.Equal(t, result, number*10)
				assert.NoError(t, err)

				return nil
			},
		)
	}

	assert.Equal(t, 10, b.Size())

	b.Reduce()

	wg.Wait()
}

func ExampleBatch() {
	var wg sync.WaitGroup
	wg.Add(2)

	b := NewBatch(
		context.Background(), func(input []int) []int {
			fmt.Println(input)
			return input
		},
	)

	ContinueWithNoResult(
		context.Background(), b.Append(1), func(_ context.Context, result int, err error) error {
			wg.Done()

			return nil
		},
	)

	ContinueWithNoResult(
		context.Background(), b.Append(2), func(_ context.Context, result int, err error) error {
			wg.Done()

			return nil
		},
	)

	b.Reduce()

	wg.Wait()

	// Output:
	// [1 2]
}
