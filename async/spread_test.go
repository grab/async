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
