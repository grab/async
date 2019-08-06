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

// Code generated by mockery v1.0.0. DO NOT EDIT.
package mocks

import (
	"github.com/stretchr/testify/mock"
	"gitlab.myteksi.net/grab-x/async"
)

// Partitioner is an autogenerated mock type for the Partitioner type
type Partitioner struct {
	mock.Mock
}

// Append provides a mock function with given fields: items
func (_m *Partitioner) Append(items interface{}) async.Task {
	ret := _m.Called(items)

	var r0 async.Task
	if rf, ok := ret.Get(0).(func(interface{}) async.Task); ok {
		r0 = rf(items)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(async.Task)
		}
	}

	return r0
}

// Partition provides a mock function with given fields:
func (_m *Partitioner) Partition() map[string][]interface{} {
	ret := _m.Called()

	var r0 map[string][]interface{}
	if rf, ok := ret.Get(0).(func() map[string][]interface{}); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string][]interface{})
		}
	}

	return r0
}