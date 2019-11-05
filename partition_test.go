// Copyright 2019 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPartitioner(t *testing.T) {
	partitionFunc := func(data interface{}) (string, bool) {
		xevent, ok := data.(map[string]string)
		if !ok {
			return "", false
		}
		key, ok := xevent["pre"]
		return key, ok
	}
	p := NewPartitioner(context.Background(), partitionFunc)

	input1 := []interface{}{
		map[string]string{"pre": "a", "val": "val1"},
		map[string]string{"pre": "b", "val": "val2"},
		map[string]string{"pre": "a", "val": "val4"},
		map[string]string{"pre": "c", "val": "val5"},
	}

	input2 := []interface{}{
		map[string]string{"pre": "a", "val": "val3"},
		map[string]string{"pre": "c", "val": "val4"},
	}

	expectedRes1 := map[string][]interface{}{
		"a": {
			map[string]string{"pre": "a", "val": "val1"},
			map[string]string{"pre": "a", "val": "val4"},
			map[string]string{"pre": "a", "val": "val3"},
		},
		"b": {
			map[string]string{"pre": "b", "val": "val2"},
		},
		"c": {
			map[string]string{"pre": "c", "val": "val5"},
			map[string]string{"pre": "c", "val": "val4"},
		},
	}

	expectedRes2 := map[string][]interface{}{
		"a": {
			map[string]string{"pre": "a", "val": "val3"},
			map[string]string{"pre": "a", "val": "val1"},
			map[string]string{"pre": "a", "val": "val4"},
		},
		"b": {
			map[string]string{"pre": "b", "val": "val2"},
		},
		"c": {
			map[string]string{"pre": "c", "val": "val4"},
			map[string]string{"pre": "c", "val": "val5"},
		},
	}

	t1 := p.Append(input1)
	t2 := p.Append(input2)
	_, _ = t1.Outcome()
	_, _ = t2.Outcome()

	res := p.Partition()
	assert.True(t, reflect.DeepEqual(expectedRes1, res) || reflect.DeepEqual(expectedRes2, res))
}

func ExamplePartitioner() {
	partitionFunc := func(data interface{}) (string, bool) {
		xevent, ok := data.(map[string]string)
		if !ok {
			return "", false
		}
		key, ok := xevent["pre"]
		return key, ok
	}
	p := NewPartitioner(context.Background(), partitionFunc)

	input := []interface{}{
		map[string]string{"pre": "a", "val": "val1"},
		map[string]string{"pre": "b", "val": "val2"},
		map[string]string{"pre": "a", "val": "val4"},
		map[string]string{"pre": "c", "val": "val5"},
	}
	t := p.Append(input)
	_, _ = t.Outcome()

	res := p.Partition()
	first := res["a"][0].(map[string]string)
	fmt.Println(first["pre"])
	fmt.Println(first["val"])

	// Output:
	// a
	// val1
}

func TestQueue(t *testing.T) {
	q := newQueue()
	input1 := partitionedItems{
		"a": []interface{}{"val1"},
		"b": []interface{}{"val2"},
	}

	input2 := partitionedItems{
		"a": []interface{}{"val4"},
		"c": []interface{}{"val5"},
	}

	expectedRes := []partitionedItems{
		{
			"a": []interface{}{"val1"},
			"b": []interface{}{"val2"},
		}, {
			"a": []interface{}{"val4"},
			"c": []interface{}{"val5"},
		},
	}

	q.Append(input1)
	q.Append(input2)
	assert.Equal(t, expectedRes, q.Flush())
}

func TestQuery_flush(t *testing.T) {
	q := newQueue()

	// fill greater than default capacity
	items := defaultCapacity + 10
	for x := 0; x < items; x++ {
		q.Append(partitionedItems{})
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
