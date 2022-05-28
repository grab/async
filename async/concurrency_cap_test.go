// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunWithConcurrencyLevelC_HappyPath(t *testing.T) {
	tests := []struct {
		desc        string
		taskCount   int
		concurrency int
	}{
		{
			desc:        "10 tasks in channel to be run with default concurrency",
			taskCount:   10,
			concurrency: 0,
		},
		{
			desc:        "10 tasks in channel to be run with 2 workers",
			taskCount:   10,
			concurrency: 2,
		},
		{
			desc:        "10 tasks in channel to be run with 10 workers",
			taskCount:   10,
			concurrency: 10,
		},
		{
			desc:        "10 tasks in channel to be run with 20 workers",
			taskCount:   10,
			concurrency: 20,
		},
	}

	for _, test := range tests {
		m := test

		resChan := make(chan struct{}, m.taskCount)
		taskChan := make(chan Task[struct{}])

		go func() {
			defer close(taskChan)

			for i := 0; i < m.taskCount; i++ {
				taskChan <- NewTask(
					func(context.Context) (struct{}, error) {
						resChan <- struct{}{}
						time.Sleep(time.Millisecond * 10)

						return struct{}{}, nil
					},
				)
			}
		}()

		p := RunWithConcurrencyLevelC(context.Background(), m.concurrency, taskChan)
		err := p.Error()
		close(resChan)

		assert.Nil(t, err, m.desc)

		var res []struct{}
		for r := range resChan {
			res = append(res, r)
		}

		assert.Len(t, res, m.taskCount, m.desc)
	}
}

func TestRunWithConcurrencyLevelC_SadPath(t *testing.T) {
	tests := []struct {
		desc        string
		taskCount   int
		concurrency int
		timeOut     time.Duration // in millisecond
	}{
		{
			desc:        "2 workers cannot finish 10 tasks in 20 ms where 1 task takes 10 ms. Context cancelled while waiting for available worker",
			taskCount:   10,
			concurrency: 2,
			timeOut:     20,
		},
		{
			desc:        "once 10 tasks are completed, workers will wait for more task. Then context will timeout in 20ms",
			taskCount:   10,
			concurrency: 20,
			timeOut:     20,
		},
	}

	for _, test := range tests {
		m := test

		taskChan := make(chan SilentTask, m.taskCount)
		ctx, _ := context.WithTimeout(context.Background(), m.timeOut*time.Millisecond)

		go func() {
			for i := 0; i < m.taskCount; i++ {
				taskChan <- NewSilentTask(
					func(context.Context) error {
						time.Sleep(time.Millisecond * 10)

						return nil
					},
				)
			}
		}()

		st := RunWithConcurrencyLevelC(ctx, m.concurrency, taskChan)
		err := st.Error()

		assert.NotNil(t, err, m.desc)
	}
}

func TestRunWithConcurrencyLevelC_VerifyTaskDrainingOnCancel(t *testing.T) {
	taskChan := make(chan SilentTask, 6)
	tasks := make([]SilentTask, 6)

	for i := 0; i < 6; i++ {
		tasks[i] = NewSilentTask(
			func(context.Context) error {
				time.Sleep(time.Millisecond * 50)

				return nil
			},
		)

		taskChan <- tasks[i]
	}

	ctxWithCancel, cancel := context.WithCancel(context.Background())
	cancel()

	RunWithConcurrencyLevelC(ctxWithCancel, 2, taskChan)

	// Pause to wait for draining to complete
	time.Sleep(time.Millisecond * 50)

	// Remaining tasks should be cancelled
	for i := 0; i < 6; i++ {
		assert.Equal(t, IsCancelled, tasks[i].State())
	}
}

func TestRunWithConcurrencyLevelS(t *testing.T) {
	resChan := make(chan int, 6)
	works := make([]Work[struct{}], 6)

	for i := range works {
		j := i

		works[j] = func(context.Context) (struct{}, error) {
			resChan <- j / 2
			time.Sleep(time.Millisecond * 10)

			return struct{}{}, nil
		}
	}

	tasks := NewTasks(works...)
	RunWithConcurrencyLevelS(context.Background(), 2, tasks)

	WaitAll(tasks)
	close(resChan)

	var res []int
	for r := range resChan {
		res = append(res, r)
	}

	assert.Equal(t, []int{0, 0, 1, 1, 2, 2}, res)
}

func TestTestRunWithConcurrencyLevelS_WithCancellingHalfway(t *testing.T) {
	tasks := make([]SilentTask, 6)

	for i := range tasks {
		j := i

		tasks[j] = NewSilentTask(
			func(context.Context) error {
				time.Sleep(time.Millisecond * 10)

				return nil
			},
		)
	}

	ctxWithCancel, cancel := context.WithCancel(context.Background())

	RunWithConcurrencyLevelS(ctxWithCancel, 2, tasks)

	// Sleep and cancel right after first 2 tasks complete
	time.Sleep(time.Millisecond * 15)
	cancel()

	WaitAll(tasks)

	// Remaining tasks should be cancelled
	for i := 2; i < 6; i++ {
		assert.Equal(t, IsCancelled, tasks[i].State())
	}
}

func TestRunWithConcurrencyLevelS_WithZeroConcurrency(t *testing.T) {
	resChan := make(chan int, 6)
	works := make([]Work[struct{}], 6)

	for i := range works {
		j := i

		works[j] = func(context.Context) (struct{}, error) {
			resChan <- 1
			time.Sleep(time.Millisecond * 10)

			return struct{}{}, nil
		}
	}

	tasks := NewTasks(works...)
	RunWithConcurrencyLevelS(context.Background(), 0, tasks)

	WaitAll(tasks)
	close(resChan)

	var res []int
	for r := range resChan {
		res = append(res, r)
	}

	assert.Equal(t, []int{1, 1, 1, 1, 1, 1}, res)
}

func ExampleRunWithConcurrencyLevelS() {
	resChan := make(chan int, 6)
	works := make([]Work[struct{}], 6, 6)

	for i := range works {
		j := i

		works[j] = func(context.Context) (struct{}, error) {
			fmt.Println(j / 2)
			time.Sleep(time.Millisecond * 10)

			return struct{}{}, nil
		}
	}

	tasks := NewTasks(works...)
	RunWithConcurrencyLevelS(context.Background(), 2, tasks)

	WaitAll(tasks)
	close(resChan)

	var res []int
	for r := range resChan {
		res = append(res, r)
	}

	// Output:
	// 0
	// 0
	// 1
	// 1
	// 2
	// 2
}
