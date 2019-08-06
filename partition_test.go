// Copyright (c) 2012-2019 Grabtaxi Holdings PTE LTD (GRAB), All Rights Reserved. NOTICE: All information contained herein
// is, and remains the property of GRAB. The intellectual and technical concepts contained herein are confidential, proprietary
// and controlled by GRAB and may be covered by patents, patents in process, and are protected by trade secret or copyright law.
//
// You are strictly forbidden to copy, download, store (in any medium), transmit, disseminate, adapt or change this material
// in any way unless prior written permission is obtained from GRAB. Access to the source code contained herein is hereby
// forbidden to anyone except current GRAB employees or contractors with binding Confidentiality and Non-disclosure agreements
// explicitly covering such access.
//
// The copyright notice above does not evidence any actual or intended publication or disclosure of this source code,
// which includes information that is confidential and/or proprietary, and is a trade secret, of GRAB.
//
// ANY REPRODUCTION, MODIFICATION, DISTRIBUTION, PUBLIC PERFORMANCE, OR PUBLIC DISPLAY OF OR THROUGH USE OF THIS SOURCE
// CODE WITHOUT THE EXPRESS WRITTEN CONSENT OF GRAB IS STRICTLY PROHIBITED, AND IN VIOLATION OF APPLICABLE LAWS AND
// INTERNATIONAL TREATIES. THE RECEIPT OR POSSESSION OF THIS SOURCE CODE AND/OR RELATED INFORMATION DOES NOT CONVEY
// OR IMPLY ANY RIGHTS TO REPRODUCE, DISCLOSE OR DISTRIBUTE ITS CONTENTS, OR TO MANUFACTURE, USE, OR SELL ANYTHING
// THAT IT MAY DESCRIBE, IN WHOLE OR IN PART.

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
