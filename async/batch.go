// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import (
	"context"
	"sync"
)

type batchEntry[P any, T any] struct {
	id      uint64
	payload P // Will be used as input when the batch is processed
}

type batch[P any, T any] struct {
	sync.RWMutex
	ctx           context.Context
	lastID        uint64                  // The last id for result matching
	pending       []batchEntry[P, T]      // The task queue to be executed in one batch
	batchExecutor Task[map[uint64]T]      // The current batch executor
	batch         chan []batchEntry[P, T] // The channel to submit a batch to be processed by the above executor
	processFn     func([]P) []T           // The func which will be executed to process one batch of tasks
}

// Batch represents a batch where tasks can be appended and processed in one go.
type Batch[P any, T any] interface {
	Append(payload P) Task[T]
	Size() int
	Reduce()
}

// NewBatch creates a new batch
func NewBatch[P any, T any](ctx context.Context, processFn func([]P) []T) Batch[P, T] {
	return &batch[P, T]{
		ctx:       ctx,
		pending:   []batchEntry[P, T]{},
		batch:     make(chan []batchEntry[P, T]),
		processFn: processFn,
	}
}

// Append adds a new payload to the batch and returns the task for that particular payload.
// You should listen for the outcome, as the task will be executed by the reducer.
func (b *batch[P, T]) Append(payload P) Task[T] {
	b.Lock()
	defer b.Unlock()

	b.lastID = b.lastID + 1
	id := b.lastID

	// Make sure we have a batch executor
	if b.batchExecutor == nil {
		b.batchExecutor = b.createBatchExecutor()
	}

	// Extract result from the processed batch
	t := ContinueWith(
		b.ctx, b.batchExecutor, func(_ context.Context, batchResult map[uint64]T, _ error) (T, error) {
			return batchResult[id], nil
		},
	)

	// Add to the task queue
	b.pending = append(
		b.pending, batchEntry[P, T]{
			id:      id,
			payload: payload,
		},
	)

	return t
}

// Reduce executes all pending tasks in one batch.
func (b *batch[P, T]) Reduce() {
	b.Lock()
	defer b.Unlock()

	// Skip if the queue is empty
	if len(b.pending) == 0 {
		return
	}

	// Capture pending tasks and reset the queue
	pendingBatch := b.pending
	b.pending = []batchEntry[P, T]{}

	// Run the current batch using the existing executor
	b.batch <- pendingBatch

	// Prepare a new executor
	b.batchExecutor = b.createBatchExecutor()
}

// Size returns the length of the pending queue.
func (b *batch[P, T]) Size() int {
	b.RLock()
	defer b.RUnlock()

	return len(b.pending)
}

// createBatchExecutor creates an executor for one batch of tasks.
func (b *batch[P, T]) createBatchExecutor() Task[map[uint64]T] {
	return Invoke(
		b.ctx, func(context.Context) (map[uint64]T, error) {
			// Block here until a batch is submitted to be processed
			pendingBatch := <-b.batch

			m := make(map[uint64]T)

			// Prepare the input for the batch reduce call
			input := make([]P, len(pendingBatch))
			for i, b := range pendingBatch {
				input[i] = b.payload
			}

			// Process the input
			result := b.processFn(input)
			for i, res := range result {
				id := pendingBatch[i].id
				m[id] = res
			}

			return m, nil
		},
	)
}
