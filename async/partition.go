// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import (
	"context"
	"sync"
)

type partitioner[K comparable, V any] struct {
	sync.RWMutex
	ctx         context.Context
	queue       *queue[K, V]
	partitionFn PartitionFunc[K, V]
}

const defaultCapacity = 1 << 14

type partitionedItems[K comparable, V any] map[K][]V

// Partitioner divides items into separate partitions.
type Partitioner[K comparable, V any] interface {
	// Append items to the partition queue.
	Append(items ...V) SilentTask

	// Partition divides items into separate partitions.
	Partition() map[K][]V
}

// PartitionFunc takes in data and then returns a key and whether the
// key can be used to route data into a partition.
type PartitionFunc[K comparable, V any] func(data V) (key K, ok bool)

// NewPartitioner creates a new partitioner.
func NewPartitioner[K comparable, V any](ctx context.Context, partitionFn PartitionFunc[K, V]) Partitioner[K, V] {
	return &partitioner[K, V]{
		ctx:         ctx,
		queue:       newQueue[K, V](),
		partitionFn: partitionFn,
	}
}

func (p *partitioner[K, V]) Append(items ...V) SilentTask {
	return InvokeInSilence(
		p.ctx, func(context.Context) error {
			p.queue.Append(p.transform(items))

			return nil
		},
	)
}

func (p *partitioner[K, V]) transform(items []V) partitionedItems[K, V] {
	mapped := partitionedItems[K, V]{}
	for _, item := range items {
		if key, ok := p.partitionFn(item); ok {
			mapped[key] = append(mapped[key], item)
		}
	}

	return mapped
}

func (p *partitioner[K, V]) Partition() map[K][]V {
	out := partitionedItems[K, V]{}

	for _, pMap := range p.queue.Flush() {
		for k, v := range pMap {
			out[k] = append(out[k], v...)
		}
	}

	return out
}

// ------------------------------------------------------

// Queue represents a queue that supports concurrent inserts.
type queue[K comparable, V any] struct {
	sync.Mutex
	queue []partitionedItems[K, V]
}

// newQueue creates a new event queue.
func newQueue[K comparable, V any]() *queue[K, V] {
	return &queue[K, V]{
		queue: make([]partitionedItems[K, V], 0, defaultCapacity),
	}
}

// Append appends to the concurrent queue.
func (q *queue[K, V]) Append(events partitionedItems[K, V]) {
	q.Lock()
	defer q.Unlock()

	q.queue = append(q.queue, events)
}

// Flush flushes the queue.
func (q *queue[K, V]) Flush() []partitionedItems[K, V] {
	q.Lock()
	defer q.Unlock()

	flushed := q.queue
	q.queue = make([]partitionedItems[K, V], 0, defaultCapacity)

	return flushed
}
