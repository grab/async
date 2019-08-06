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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func newTasks() []Task {
	work := func(context.Context) (interface{}, error) {
		return 1, nil
	}

	return NewTasks(work, work, work, work, work)
}

func TestThrottle(t *testing.T) {
	tasks := newTasks()

	// Throttle and calculate the duration
	t0 := time.Now()
	task := Throttle(context.Background(), tasks, 3, 50*time.Millisecond)
	_, _ = task.Outcome() // Wait

	// Make sure we completed within duration
	dt := int(time.Now().Sub(t0).Seconds() * 1000)
	assert.True(t, dt > 50 && dt < 100, fmt.Sprintf("%v ms.", dt))
}

func TestThrottle_Cancel(t *testing.T) {
	tasks := newTasks()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Throttle and calculate the duration
	Throttle(ctx, tasks, 3, 50*time.Millisecond)
	WaitAll(tasks)
	cancelled := 0
	for _, task := range tasks {
		if task.State() == IsCancelled {
			cancelled++
		}
	}

	assert.Equal(t, 5, cancelled)
}

func TestSpread(t *testing.T) {
	tasks := newTasks()
	within := 200 * time.Millisecond

	// Spread and calculate the duration
	t0 := time.Now()
	task := Spread(context.Background(), within, tasks)
	_, _ = task.Outcome() // Wait

	// Make sure we completed within duration
	dt := int(time.Now().Sub(t0).Seconds() * 1000)
	assert.True(t, dt > 150 && dt < 250, fmt.Sprintf("%v ms.", dt))

	// Make sure all tasks are done
	for _, task := range tasks {
		v, _ := task.Outcome()
		assert.Equal(t, 1, v.(int))
	}
}

func ExampleSpread() {
	tasks := newTasks()
	within := 200 * time.Millisecond

	// Spread
	task := Spread(context.Background(), within, tasks)
	_, _ = task.Outcome() // Wait

	// Make sure all tasks are done
	for _, task := range tasks {
		v, _ := task.Outcome()
		fmt.Println(v)
	}

	// Output:
	// 1
	// 1
	// 1
	// 1
	// 1

}
