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

func TestRepeat(t *testing.T) {
	assert.NotPanics(t, func() {
		out := make(chan bool, 1)
		task := Repeat(context.TODO(), time.Nanosecond*10, func(context.Context) (interface{}, error) {
			out <- true
			return nil, nil
		})

		<-out
		v := <-out
		assert.True(t, v)
		task.Cancel()
	})
}

func ExampleRepeat() {
	out := make(chan bool, 1)
	task := Repeat(context.TODO(), time.Nanosecond*10, func(context.Context) (interface{}, error) {
		out <- true
		return nil, nil
	})

	<-out
	v := <-out
	fmt.Println(v)
	task.Cancel()

	// Output:
	// true
}

/*
func TestRepeatFirstActionPanic(t *testing.T) {
	assert.NotPanics(t, func() {
		task := Repeat(context.TODO(), time.Nanosecond*10, func(context.Context) (interface{}, error) {
			panic("test")
		})

		task.Cancel()
	})
}

func TestRepeatPanic(t *testing.T) {
	assert.NotPanics(t, func() {
		var counter int32
		task := Repeat(context.TODO(), time.Nanosecond*10, func(context.Context) (interface{}, error) {
			atomic.AddInt32(&counter, 1)
			panic("test")
		})

		for atomic.LoadInt32(&counter) <= 10 {
		}

		task.Cancel()
	})
}
*/
