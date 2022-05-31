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

func TestThrottle(t *testing.T) {
	tasks := newTasks()

	// Throttle and calculate the duration
	t0 := time.Now()

	throttledTask := Throttle(context.Background(), tasks, 3, 50*time.Millisecond)
	throttledTask.Wait()

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
	for _, t := range tasks {
		if t.State() == IsCancelled {
			cancelled++
		}
	}

	assert.Equal(t, 5, cancelled)
}
