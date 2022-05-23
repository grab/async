// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type animal struct {
	species string
	name    string
}

func TestPartitioner(t *testing.T) {
	partitionFunc := func(data animal) (string, bool) {
		if data.species == "" {
			return "", false
		}

		return data.species, true
	}

	p := NewPartitioner(context.Background(), partitionFunc)

	input1 := []animal{
		{"dog", "name1"},
		{"snail", "name2"},
		{"dog", "name4"},
		{"cat", "name5"},
	}

	input2 := []animal{
		{"dog", "name3"},
		{"cat", "name4"},
	}

	t1 := p.Append(input1...)
	t2 := p.Append(input2...)
	t1.Wait()
	t2.Wait()

	expected1 := map[string][]animal{
		"dog": {
			{"dog", "name1"},
			{"dog", "name4"},
			{"dog", "name3"},
		},
		"snail": {
			{"snail", "name2"},
		},
		"cat": {
			{"cat", "name5"},
			{"cat", "name4"},
		},
	}

	expected2 := map[string][]animal{
		"dog": {
			{"dog", "name3"},
			{"dog", "name1"},
			{"dog", "name4"},
		},
		"snail": {
			{"snail", "name2"},
		},
		"cat": {
			{"cat", "name4"},
			{"cat", "name5"},
		},
	}

	res := p.Partition()
	assert.True(t, reflect.DeepEqual(expected1, res) || reflect.DeepEqual(expected2, res))
}

func ExamplePartitioner() {
	partitionFunc := func(data animal) (string, bool) {
		if data.species == "" {
			return "", false
		}

		return data.species, true
	}

	p := NewPartitioner(context.Background(), partitionFunc)

	input := []animal{
		{"dog", "name1"},
		{"snail", "name2"},
		{"dog", "name4"},
		{"cat", "name5"},
	}

	t := p.Append(input...)
	t.Wait()

	res := p.Partition()
	first := res["dog"]
	fmt.Println(first[0])
	fmt.Println(first[1])

	// Output:
	// {dog name1}
	// {dog name4}
}

func TestQueue(t *testing.T) {
	q := newQueue[string, string]()
	input1 := partitionedItems[string, string]{
		"a": []string{"val1"},
		"b": []string{"val2"},
	}

	input2 := partitionedItems[string, string]{
		"a": []string{"val4"},
		"c": []string{"val5"},
	}

	expectedRes := []partitionedItems[string, string]{
		{
			"a": []string{"val1"},
			"b": []string{"val2"},
		}, {
			"a": []string{"val4"},
			"c": []string{"val5"},
		},
	}

	q.Append(input1)
	q.Append(input2)

	assert.Equal(t, expectedRes, q.Flush())
}

func TestQuery_flush(t *testing.T) {
	q := newQueue[string, string]()

	// fill greater than default capacity
	items := defaultCapacity + 10
	for x := 0; x < items; x++ {
		q.Append(partitionedItems[string, string]{})
	}

	assert.True(t, defaultCapacity < cap(q.queue))
	assert.True(t, defaultCapacity < len(q.queue))

	// flush
	flushedItems := q.Flush()

	// validate
	assert.Equal(t, items, len(flushedItems))
	assert.Equal(t, 0, len(q.queue))
	assert.Equal(t, defaultCapacity, cap(q.queue))
}
