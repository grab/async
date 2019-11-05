// Copyright 2019 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import (
	"context"
	"reflect"
	"sync"
)

type partitioner struct {
	sync.RWMutex
	ctx       context.Context
	queue     *queue
	partition PartitionFunc // The function which will be executed to process the items of the NewBatch
}

const defaultCapacity = 1 << 14

type partitionedItems map[string][]interface{}

// Partitioner partitions events
type Partitioner interface {
	// Append items to the queue which is pending partition
	Append(items interface{}) Task

	// Partition items and output the result
	Partition() map[string][]interface{}
}

// PartitionFunc takes in data and outputs key
// if ok is false, the data doesn't fall into and partition
type PartitionFunc func(data interface{}) (key string, ok bool)

// NewPartitioner creates a new partitioner
func NewPartitioner(ctx context.Context, partition PartitionFunc) Partitioner {
	return &partitioner{
		ctx:       ctx,
		queue:     newQueue(),
		partition: partition,
	}
}

// Append adds a batch of events to the buffer
func (p *partitioner) Append(items interface{}) Task {
	return Invoke(p.ctx, func(context.Context) (interface{}, error) {
		p.queue.Append(p.transform(items))
		return nil, nil
	})
}

// transform creates a map of scope to event
func (p *partitioner) transform(items interface{}) partitionedItems {
	t := reflect.TypeOf(items)
	if t.Kind() != reflect.Slice {
		panic("transform requires for slice")
	}

	rv := reflect.ValueOf(items)
	mapped := partitionedItems{}
	for i := 0; i < rv.Len(); i++ {
		e := rv.Index(i).Interface()
		if key, ok := p.partition(e); ok {
			mapped[key] = append(mapped[key], e)
		}
	}
	return mapped
}

// Partition flushes the list of events and clears up the buffer
func (p *partitioner) Partition() map[string][]interface{} {
	out := partitionedItems{}
	for _, pMap := range p.queue.Flush() {
		for k, v := range pMap {
			out[k] = append(out[k], v...)
		}
	}
	return out
}

// ------------------------------------------------------

// Queue represents a batch queue for faster inserts
type queue struct {
	sync.Mutex
	queue []partitionedItems
}

// newQueue creates a new event queue
func newQueue() *queue {
	return &queue{
		queue: make([]partitionedItems, 0, defaultCapacity),
	}
}

// Append appends to the concurrent queue
func (q *queue) Append(events partitionedItems) {
	q.Lock()
	q.queue = append(q.queue, events)
	q.Unlock()
}

// Flush flushes the event queue
func (q *queue) Flush() []partitionedItems {
	q.Lock()
	defer q.Unlock()

	flushed := q.queue
	q.queue = make([]partitionedItems, 0, defaultCapacity)

	return flushed
}
