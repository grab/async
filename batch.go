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
	"errors"
	"fmt"
	"sync"
)

type batchEntry struct {
	id      uint64
	payload interface{} // Will be used as input when the batch is processed
	task    Task        // The callback will be called when this entry is processed
}

type batch struct {
	sync.RWMutex
	ctx       context.Context
	lastID    uint64                            // The last id for result matching
	pending   []batchEntry                      // The pending entries to the batch
	batchTask Task                              // The current batch task
	batch     chan []batchEntry                 // The current batch channel to execute
	process   func([]interface{}) []interface{} // The function which will be executed to process the items of the NewBatch
}

// Batch represents a batch where one can append to the batch and process it as a whole.
type Batch interface {
	Append(payload interface{}) Task
	Size() int
	Reduce()
}

// NewBatch creates a new batch
func NewBatch(ctx context.Context, process func([]interface{}) []interface{}) Batch {
	return &batch{
		ctx:     ctx,
		pending: []batchEntry{},
		batch:   make(chan []batchEntry),
		process: process,
	}
}

// Append adds a new payload to the batch and returns the task for that particular
// payload. You should listen for the outcome, as the task will be executed by the reducer.
func (b *batch) Append(payload interface{}) Task {
	b.Lock()
	defer b.Unlock()

	b.lastID = b.lastID + 1
	id := b.lastID

	// Make sure we have a batch task
	if b.batchTask == nil {
		b.batchTask = b.createBatchTask()
	}

	// Batch task will need to continue with this one
	t := b.batchTask.ContinueWith(b.ctx, func(batchResult interface{}, _ error) (interface{}, error) {
		if res, ok := batchResult.(map[uint64]interface{}); ok {
			return res[id], nil
		}

		actualType := fmt.Sprintf("%T", batchResult)
		return nil, errors.New("Invalid batch type, got: " + actualType)
	})

	// Add to the task queue
	b.pending = append(b.pending, batchEntry{
		id:      id,
		payload: payload,
		task:    t,
	})

	// Return the task we created
	return t
}

// Reduce will send a batch
func (b *batch) Reduce() {
	b.Lock()
	defer b.Unlock()

	// Skip if the queue is empty
	if len(b.pending) == 0 {
		return
	}

	// Prepare the batch
	batch := append([]batchEntry{}, b.pending...)

	// Run the current batch
	b.batch <- batch

	// Swap the batch
	b.batchTask = b.createBatchTask()
}

// Size returns the length of the pending queue
func (b *batch) Size() int {
	b.RLock()
	defer b.RUnlock()
	return len(b.pending)
}

// createBatchTask creates a task for the batch. Triggering this task will trigger the whole batch.
func (b *batch) createBatchTask() Task {
	return Invoke(b.ctx, func(context.Context) (interface{}, error) {
		// block here until a batch is ordered to be processed
		batch := <-b.batch
		m := map[uint64]interface{}{}

		// prepare the input for the batch reduce call
		input := make([]interface{}, len(batch))
		for i, b := range batch {
			input[i] = b.payload
		}

		// process the input
		result := b.process(input)
		for i, res := range result {
			id := batch[i].id
			m[id] = res
		}

		// return the map of associations
		return m, nil
	})
}
