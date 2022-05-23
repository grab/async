// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestForkJoin(t *testing.T) {
	first := NewTask(
		func(context.Context) (int, error) {
			return 1, nil
		},
	)

	second := NewTask(
		func(context.Context) (interface{}, error) {
			return nil, errors.New("some error")
		},
	)

	third := NewTask(
		func(context.Context) (int, error) {
			return 3, nil
		},
	)

	ForkJoin(context.Background(), []SilentTask{first, second, third})

	outcome1, error1 := first.Outcome()
	assert.Equal(t, 1, outcome1)
	assert.Nil(t, error1)

	outcome2, error2 := second.Outcome()
	assert.Nil(t, outcome2)
	assert.NotNil(t, error2)

	outcome3, error3 := third.Outcome()
	assert.Equal(t, 3, outcome3)
	assert.Nil(t, error3)
}

func ExampleForkJoin() {
	first := NewTask(
		func(context.Context) (int, error) {
			return 1, nil
		},
	)

	second := NewTask(
		func(context.Context) (interface{}, error) {
			return nil, errors.New("some error")
		},
	)

	ForkJoin(context.Background(), []SilentTask{first, second})

	fmt.Println(first.Outcome())
	fmt.Println(second.Outcome())

	// Output:
	// 1 <nil>
	// <nil> some error
}
