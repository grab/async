// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

// Package async simplifies the implementation of orchestration patterns for concurrent systems. Currently, it includes:
//
// Asynchronous tasks with cancellations, context propagation and state.
//
// Task chaining by using continuations.
//
// Fork/join pattern - running a bunch of work and waiting for everything to finish.
//
// Throttling pattern - throttling task execution on a specified rate.
//
// Spread pattern - spreading tasks across time.
//
// Partition pattern - partitioning data concurrently.
//
// Repeat pattern - repeating a certain task at a specified interval.
//
// Batch pattern - batching many tasks into a single one with individual continuations.

package async
