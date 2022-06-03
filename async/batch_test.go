// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

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
			b.Append(number), func(_ context.Context, err error) error {
				defer wg.Done()

				assert.Nil(t, err)

				return nil
			},
		).Execute(context.Background())
	}

	assert.Equal(t, 10, b.Size())

	b.Process(context.Background())

	wg.Wait()

	for i := 0; i < taskCount; i++ {
		assert.Equal(t, i*10, <-out)
	}
}

func TestBatcher_AppendAutoProcessBySize(t *testing.T) {
	const taskCount = 10

	out := make(chan int, taskCount)

	// Processor that multiplies items by 10 all at once
	b := NewBatcher(
		func(input []int) error {
			for _, number := range input {
				out <- number * 10
			}

			return nil
		},
		WithAutoProcessSize(taskCount),
	)

	tasks := make([]SilentTask, taskCount)
	for i := 0; i < taskCount; i++ {
		number := i

		tasks[i] = ContinueInSilence(
			b.Append(number), func(_ context.Context, err error) error {
				assert.Nil(t, err)

				return err
			},
		).Execute(context.Background())
	}

	WaitAll(tasks)

	assert.Equal(t, 0, b.Size(), "All pending tasks should have been auto processed")

	for i := 0; i < taskCount; i++ {
		assert.Equal(t, i*10, <-out)
	}
}

func TestBatcher_AutoProcessOnInterval(t *testing.T) {
	const taskCount = 10

	out := make(chan int, taskCount)

	// Processor that multiplies items by 10 all at once
	b := NewBatcher(
		func(input []int) error {
			for _, number := range input {
				out <- number * 10
			}

			return nil
		},
		WithAutoProcessInterval(100*time.Millisecond),
	)

	defer b.Shutdown()

	tasks := make([]SilentTask, taskCount)
	for i := 0; i < taskCount; i++ {
		number := i

		tasks[i] = ContinueInSilence(
			b.Append(number), func(_ context.Context, err error) error {
				assert.Nil(t, err)

				return err
			},
		).Execute(context.Background())
	}

	assert.Equal(t, 10, b.Size())

	WaitAll(tasks)

	for i := 0; i < taskCount; i++ {
		assert.Equal(t, i*10, <-out)
	}
}

func TestBatcher_Shutdown(t *testing.T) {
	const taskCount = 10

	out := make(chan int, taskCount)

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
			b.Append(number), func(_ context.Context, err error) error {
				assert.Nil(t, err)

				return err
			},
		).Execute(context.Background())
	}

	assert.Equal(t, 10, b.Size())

	// Shutdown should process all pending tasks
	b.Shutdown()

	for i := 0; i < taskCount; i++ {
		assert.Equal(t, i*10, <-out)
	}
}

func TestBatcher_ShutdownWithTimeout(t *testing.T) {
	const taskCount = 10

	// Processor that multiplies items by 10 all at once
	b := NewBatcher(
		func(input []int) error {
			time.Sleep(100 * time.Millisecond)

			return nil
		},
		WithShutdownGraceDuration(50*time.Millisecond),
	)

	tasks := make([]SilentTask, taskCount)
	for i := 0; i < taskCount; i++ {
		number := i

		tasks[i] = ContinueInSilence(
			b.Append(number), func(_ context.Context, err error) error {
				return err
			},
		).Execute(context.Background())
	}

	assert.Equal(t, 10, b.Size())

	// Shutdown should process all pending tasks
	b.Shutdown()

	WaitAll(tasks)

	for i := 0; i < taskCount; i++ {
		assert.Equal(t, IsCompleted, tasks[i].State())
		assert.Equal(t, context.DeadlineExceeded, tasks[i].Error())
	}
}

func ExampleBatcher() {
	var wg sync.WaitGroup
	wg.Add(2)

	b := NewBatcher(
		func(input []int) error {
			fmt.Println(input)
			return nil
		},
	)

	ContinueInSilence(
		b.Append(1), func(_ context.Context, err error) error {
			wg.Done()

			return nil
		},
	).Execute(context.Background())

	ContinueInSilence(
		b.Append(2), func(_ context.Context, err error) error {
			wg.Done()

			return nil
		},
	).Execute(context.Background())

	b.Process(context.Background())

	wg.Wait()

	// Output:
	// [1 2]
}
