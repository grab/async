// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrBatchChannelClosed      = errors.New("input batch channel was closed")
	ErrBatchProcessorNotActive = errors.New("batch processor has already shut down")
)

type batchEntry[P any] struct {
	id      uint64
	payload P // Will be used as input when the batch is processed
}

type batcher[P any] struct {
	sync.RWMutex
	isActive      bool
	lastID        uint64               // The last id for result matching
	pending       []batchEntry[P]      // The task queue to be executed in one batch
	batchExecutor SilentTask           // The current batch executor
	batch         chan []batchEntry[P] // The channel to submit a batch to be processed by the above executor
	processFn     func([]P) error      // The func which will be executed to process one batch of tasks
}

// Batcher represents a batch processor that can accumulate tasks and execute all in one go.
// This implementation is suitable for sitting in the back and
type Batcher[P any] interface {
	// Append adds a new payload to the batch and returns a task for that particular payload.
	// Clients can block and wait for the returned task to complete before extracting result.
	Append(payload P) SilentTask
	// Size returns the length of the pending queue.
	Size() int
	// Process executes all pending tasks in one go.
	Process()
	// Shutdown notifies this batch processor to complete its work gracefully. Future calls
	// to Append will return an error immediately and Process will be a no-op.
	Shutdown()
}

// NewBatcher returns a new Batcher
func NewBatcher[P any](processFn func([]P) error) Batcher[P] {
	return &batcher[P]{
		isActive:  true,
		pending:   []batchEntry[P]{},
		batch:     make(chan []batchEntry[P]),
		processFn: processFn,
	}
}

func (b *batcher[P]) Append(payload P) SilentTask {
	b.Lock()
	defer b.Unlock()

	if !b.isActive {
		return Completed(struct{}{}, ErrBatchProcessorNotActive)
	}

	b.lastID = b.lastID + 1
	id := b.lastID

	// Make sure we have a batch executor
	if b.batchExecutor == nil {
		b.batchExecutor = b.createBatchExecutor()
	}

	// Extract result from the processed batch
	t := ContinueInSilence(
		context.Background(), b.batchExecutor, func(_ context.Context, err error) error {
			return err
		},
	)

	// Add to the task queue
	b.pending = append(
		b.pending, batchEntry[P]{
			id:      id,
			payload: payload,
		},
	)

	return t
}

func (b *batcher[P]) Process() {
	b.Lock()
	defer b.Unlock()

	b.doProcess(false)
}

func (b *batcher[P]) doProcess(isShuttingDown bool) {
	// Skip if the queue is empty
	if len(b.pending) == 0 {
		if isShuttingDown {
			close(b.batch)
		}

		return
	}

	// Capture pending tasks and reset the queue
	pendingBatch := b.pending
	b.pending = []batchEntry[P]{}

	// Run the current batch using the existing executor
	b.batch <- pendingBatch

	// Prepare a new executor
	if !isShuttingDown {
		b.batchExecutor = b.createBatchExecutor()
	}
}

func (b *batcher[P]) Size() int {
	b.RLock()
	defer b.RUnlock()

	return len(b.pending)
}

func (b *batcher[P]) Shutdown() {
	b.Lock()
	defer b.Unlock()

	b.doProcess(true)

	b.isActive = false
}

// createBatchExecutor creates an executor for one batch of tasks.
func (b *batcher[P]) createBatchExecutor() SilentTask {
	return InvokeInSilence(
		context.Background(), func(context.Context) error {
			// Block here until a batch is submitted to be processed
			pendingBatch, ok := <-b.batch
			if !ok {
				return ErrBatchChannelClosed
			}

			// Prepare the input for the batch process call
			input := make([]P, len(pendingBatch))
			for i, b := range pendingBatch {
				input[i] = b.payload
			}

			return b.processFn(input)
		},
	)
}
