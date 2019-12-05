// Copyright 2019 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProcessTaskPool_HappyPath(t *testing.T) {
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
		taskChan := make(chan Task)

		go func() {
			for i := 0; i < m.taskCount; i++ {
				taskChan <- NewTask(func(context.Context) (interface{}, error) {
					resChan <- struct{}{}
					time.Sleep(time.Millisecond * 10)
					return nil, nil
				})
			}
			close(taskChan)
		}()
		p := Consume(context.Background(), m.concurrency, taskChan)
		_, err := p.Outcome()
		close(resChan)
		assert.Nil(t, err, m.desc)

		var res []struct{}
		for r := range resChan {
			res = append(res, r)
		}
		assert.Len(t, res, m.taskCount, m.desc)
	}
}

// test context cancellation
func TestProcessTaskPool_SadPath(t *testing.T) {
	tests := []struct {
		desc        string
		taskCount   int
		concurrency int
		timeOut     time.Duration //in millisecond
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
		taskChan := make(chan Task)
		ctx, _ := context.WithTimeout(context.Background(), m.timeOut*time.Millisecond)
		go func() {
			for i := 0; i < m.taskCount; i++ {
				taskChan <- NewTask(func(context.Context) (interface{}, error) {
					time.Sleep(time.Millisecond * 10)
					return nil, nil
				})
			}
		}()
		p := Consume(ctx, m.concurrency, taskChan)
		_, err := p.Outcome()
		assert.NotNil(t, err, m.desc)
	}
}
