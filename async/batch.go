// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrBatchProcessorNotActive = errors.New("batch processor has already shut down")
)

type batchEntry[P any] struct {
	id      uint64
	payload P // Will be used as input when the batch is processed
}

type batcher[P any] struct {
	sync.RWMutex
	*batcherConfigs
	isActive      bool
	batchID       uint64               // The current batch ID
	pending       []batchEntry[P]      // The task queue to be executed in one batch
	batchExecutor SilentTask           // The current batch executor
	batch         chan []batchEntry[P] // The channel to submit a batch to be processed by the above executor
	processFn     func([]P) error      // The func which will be executed to process one batch of tasks
}

// Batcher is a batch processor which is suitable for sitting in the back to accumulate tasks
// and then execute all in one go.
type Batcher[P any] interface {
	// Append adds a new payload to the batch and returns a task for that particular payload.
	// Clients MUST execute the returned task before blocking and waiting for it to complete
	// to extract result.
	Append(payload P) SilentTask
	// Size returns the length of the pending queue.
	Size() int
	// Process executes all pending tasks in one go.
	Process(ctx context.Context)
	// Shutdown notifies this batch processor to complete its work gracefully. Future calls
	// to Append will return an error immediately and Process will be a no-op. This is a
	// blocking call which will wait up to the configured amount of time for the last batch
	// to complete.
	Shutdown()
}

// NewBatcher returns a new Batcher
func NewBatcher[P any](processFn func([]P) error, options ...BatcherOption) Batcher[P] {
	b := &batcher[P]{
		batcherConfigs: &batcherConfigs{},
		isActive:       true,
		pending:        []batchEntry[P]{},
		batch:          make(chan []batchEntry[P], 1),
		processFn:      processFn,
	}

	for _, o := range options {
		o(b.batcherConfigs)
	}

	if b.isPeriodicAutoProcessingConfigured() {
		go func() {
			for {
				curBatchId := b.batchID

				<-time.After(b.autoProcessInterval)

				// Best effort to prevent timer from acquiring lock unnecessarily, no guarantee
				if curBatchId == b.batchID {
					func() {
						b.Lock()
						defer b.Unlock()

						b.doProcess(context.Background(), false, curBatchId)
					}()
				}

				if !b.isActive {
					return
				}
			}
		}()
	}

	return b
}

func (b *batcher[P]) isPeriodicAutoProcessingConfigured() bool {
	return b.autoProcessInterval > 0
}

func (b *batcher[P]) Append(payload P) SilentTask {
	b.Lock()
	defer b.Unlock()

	if !b.isActive {
		return Completed(struct{}{}, ErrBatchProcessorNotActive)
	}

	// Make sure we have a batch executor
	if b.batchExecutor == nil {
		b.batchExecutor = b.createBatchExecutor()
	}

	// Add to the task queue
	b.pending = append(
		b.pending, batchEntry[P]{
			payload: payload,
		},
	)

	// Auto process if configured and reached the threshold
	if b.autoProcessSize > 0 && len(b.pending) == b.autoProcessSize {
		curBatchId := b.batchID

		go func() {
			b.Lock()
			defer b.Unlock()

			b.doProcess(context.Background(), false, curBatchId)
		}()
	}

	// Extract result from the processed batch
	curBatchExecutor := b.batchExecutor

	return NewSilentTask(
		func(ctx context.Context) error {
			curBatchExecutor.Wait()
			return curBatchExecutor.Error()
		},
	)
}

func (b *batcher[P]) Size() int {
	b.RLock()
	defer b.RUnlock()

	return len(b.pending)
}

func (b *batcher[P]) Process(ctx context.Context) {
	b.Lock()
	defer b.Unlock()

	b.doProcess(ctx, false, b.batchID)
}

func (b *batcher[P]) Shutdown() {
	b.Lock()
	defer b.Unlock()

	ctx := context.Background()
	if b.shutdownGraceDuration > 0 {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, b.shutdownGraceDuration)
		defer cancel()

		ctx = ctxWithTimeout
	}

	b.doProcess(ctx, true, b.batchID)

	b.isActive = false
}

func (b *batcher[P]) doProcess(ctx context.Context, isShuttingDown bool, toProcessBatchID uint64) {
	if b.batchID != toProcessBatchID {
		return
	}

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
	b.batchExecutor.Execute(ctx)

	// Block and wait for the last batch to complete on shutting down
	if isShuttingDown {
		b.batchExecutor.Wait()
		return
	}

	// Prepare a new executor
	b.batchExecutor = b.createBatchExecutor()

	// Increment batch ID to stop the timer from processing old batch
	b.batchID += 1
}

// createBatchExecutor creates an executor for one batch of tasks.
func (b *batcher[P]) createBatchExecutor() SilentTask {
	return NewSilentTask(
		func(context.Context) error {
			// Block here until a batch is submitted to be processed
			pendingBatch := <-b.batch

			// Prepare the input for the batch process call
			input := make([]P, len(pendingBatch))
			for i, b := range pendingBatch {
				input[i] = b.payload
			}

			return b.processFn(input)
		},
	)
}
