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
