// Copyright 2019 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
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

	// reducer that multiplies items by 10 at once
	r := NewBatch(context.Background(), func(input []interface{}) []interface{} {
		result := make([]interface{}, len(input))
		for i, number := range input {
			result[i] = number.(int) * 10
		}
		return result
	})

	for i := 0; i < taskCount; i++ {
		number := i
		r.Append(number).ContinueWith(context.TODO(), func(result interface{}, err error) (interface{}, error) {
			assert.Equal(t, result.(int), number*10)
			assert.NoError(t, err)
			wg.Done()
			return nil, nil
		})
	}

	assert.Equal(t, 10, r.Size())

	r.Reduce()
	wg.Wait()
}

func ExampleBatch() {
	var wg sync.WaitGroup
	wg.Add(2)

	r := NewBatch(context.Background(), func(input []interface{}) []interface{} {
		fmt.Println(input)
		return input
	})

	r.Append(1).ContinueWith(context.TODO(), func(result interface{}, err error) (interface{}, error) {
		wg.Done()
		return nil, nil
	})
	r.Append(2).ContinueWith(context.TODO(), func(result interface{}, err error) (interface{}, error) {
		wg.Done()
		return nil, nil
	})
	r.Reduce()
	wg.Wait()

	// Output:
	// [1 2]
}
