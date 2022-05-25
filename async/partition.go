// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import (
	"context"
	"sync"

	"github.com/jamestrandung/go-data-structure/emap"
)

type partitionedItems[K comparable, V any] map[K][]V

// Partitioner divides items into separate partitions.
type Partitioner[K comparable, V any] interface {
	// Take items and divide them into separate partitions asynchronously.
	Take(items ...V) SilentTask
	// Outcome returns items divided into separate partitions.
	Outcome() map[K][]V
}

// PartitionFunc takes in data and then returns a key and whether the
// key can be used to route data into a partition.
type PartitionFunc[K comparable, V any] func(data V) (key K, ok bool)

type partitioner[K comparable, V any] struct {
	sync.Mutex
	ctx         context.Context
	partitionFn PartitionFunc[K, V]
	partitions  emap.ConcurrentMap[K, []V]
}

// NewPartitioner creates a new partitioner.
func NewPartitioner[K comparable, V any](ctx context.Context, partitionFn PartitionFunc[K, V]) Partitioner[K, V] {
	return &partitioner[K, V]{
		ctx:         ctx,
		partitions:  emap.NewConcurrentMap[K, []V](),
		partitionFn: partitionFn,
	}
}

func (p *partitioner[K, V]) Take(rawItems ...V) SilentTask {
	return InvokeInSilence(
		p.ctx, func(context.Context) error {
			partitionedItems := p.divideIntoPartitions(rawItems)

			for key, items := range partitionedItems {
				p.partitions.GetAndSetIf(
					key, func(currentItems []V, found bool) ([]V, bool) {
						if !found {
							return items, true
						}

						newItems := make([]V, len(currentItems)+len(items))
						copy(newItems, currentItems)
						copy(newItems[len(currentItems):], items)

						return newItems, true
					},
				)
			}

			return nil
		},
	)
}

func (p *partitioner[K, V]) divideIntoPartitions(items []V) partitionedItems[K, V] {
	mapped := partitionedItems[K, V]{}
	for _, item := range items {
		if key, ok := p.partitionFn(item); ok {
			mapped[key] = append(mapped[key], item)
		}
	}

	return mapped
}

func (p *partitioner[K, V]) Outcome() map[K][]V {
	out := p.partitions.AsMap()
	p.partitions = emap.NewConcurrentMap[K, []V]()

	return out
}
