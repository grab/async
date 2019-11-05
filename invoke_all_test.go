// Copyright 2019 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInvokeAll(t *testing.T) {
	resChan := make(chan int, 6)
	works := make([]Work, 6, 6)
	for i := range works {
		j := i
		works[j] = func(context.Context) (interface{}, error) {
			resChan <- j / 2
			time.Sleep(time.Millisecond * 10)
			return nil, nil
		}
	}
	tasks := NewTasks(works...)
	InvokeAll(context.Background(), 2, tasks)
	WaitAll(tasks)
	close(resChan)
	res := []int{}
	for r := range resChan {
		res = append(res, r)
	}
	assert.Equal(t, []int{0, 0, 1, 1, 2, 2}, res)
}

func TestInvokeAllWithZeroConcurrency(t *testing.T) {
	resChan := make(chan int, 6)
	works := make([]Work, 6, 6)
	for i := range works {
		j := i
		works[j] = func(context.Context) (interface{}, error) {
			resChan <- 1
			time.Sleep(time.Millisecond * 10)
			return nil, nil
		}
	}
	tasks := NewTasks(works...)
	InvokeAll(context.Background(), 0, tasks)
	WaitAll(tasks)
	close(resChan)
	res := []int{}
	for r := range resChan {
		res = append(res, r)
	}
	assert.Equal(t, []int{1, 1, 1, 1, 1, 1}, res)
}

func ExampleInvokeAll() {
	resChan := make(chan int, 6)
	works := make([]Work, 6, 6)
	for i := range works {
		j := i
		works[j] = func(context.Context) (interface{}, error) {
			fmt.Println(j / 2)
			time.Sleep(time.Millisecond * 10)
			return nil, nil
		}
	}
	tasks := NewTasks(works...)
	InvokeAll(context.Background(), 2, tasks)
	WaitAll(tasks)
	close(resChan)
	res := []int{}
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
