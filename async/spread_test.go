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

func newTasks() []Task[int] {
	work := func(context.Context) (int, error) {
		return 1, nil
	}

	return NewTasks(work, work, work, work, work)
}

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

func TestSpread(t *testing.T) {
	tasks := newTasks()
	within := 200 * time.Millisecond

	// Spread and calculate the duration
	t0 := time.Now()

	spreadTask := Spread(context.Background(), tasks, within)
	spreadTask.Wait()

	// Make sure we completed within duration
	dt := int(time.Now().Sub(t0).Seconds() * 1000)
	assert.True(t, dt > 150 && dt < 250, fmt.Sprintf("%v ms.", dt))

	// Make sure all tasks are done
	for _, task := range tasks {
		v, _ := task.Outcome()
		assert.Equal(t, 1, v)
	}
}

func ExampleSpread() {
	tasks := newTasks()
	within := 200 * time.Millisecond

	// Spread
	task := Spread(context.Background(), tasks, within)
	task.Wait()

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
