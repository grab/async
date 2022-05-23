// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync/atomic"
	"time"
)

var errCancelled = errors.New("context canceled")

var now = time.Now

// SilentWork represents a handler to execute in silence like
// background works that return no values.
type SilentWork func(context.Context) error

// Work represents a handler to execute that is expected to
// return a value of a particular type.
type Work[T any] func(context.Context) (T, error)

// State represents the state enumeration for a task.
type State byte

// Various task states.
const (
	IsCreated   State = iota // IsCreated represents a newly created task
	IsRunning                // IsRunning represents a task which is currently running
	IsCompleted              // IsCompleted represents a task which was completed successfully or errored out
	IsCancelled              // IsCancelled represents a task which was cancelled or has timed out
)

type signal chan struct{}

// SilentTask represents a unit of work to complete in silence
// like background works that return no values.
type SilentTask interface {
	// Execute starts this task asynchronously.
	Execute(ctx context.Context) SilentTask
	// Wait waits for this task to complete.
	Wait()
	// Cancel changes the state of this task to `Cancelled`.
	Cancel()
	// Error returns the error that occurred when this task was executed.
	Error() error
	// State returns the current state of this task. This operation is non-blocking.
	State() State
	// Duration returns the duration of this task.
	Duration() time.Duration
}

// Task represents a unit of work that is expected to return
// a value of a particular type.
type Task[T any] interface {
	SilentTask
	// Run starts this task asynchronously.
	Run(ctx context.Context) Task[T]
	// Outcome waits for this task to complete and returns the final result & error.
	Outcome() (T, error)
}

type outcome[T any] struct {
	result T
	err    error
}

type task[T any] struct {
	state    int32         // The current async.State of this task
	cancel   signal        // The channel for cancelling this task
	done     signal        // The channel for indicating this task has completed
	action   Work[T]       // The work to do
	outcome  outcome[T]    // This is used to store the outcome of this task
	duration time.Duration // The duration of this task, in nanoseconds
}

// NewTask creates a new task.
func NewTask[T any](action Work[T]) Task[T] {
	return &task[T]{
		action: action,
		done:   make(signal, 1),
		cancel: make(signal, 1),
	}
}

// NewTasks creates a group of new tasks.
func NewTasks[T any](actions ...Work[T]) []Task[T] {
	tasks := make([]Task[T], 0, len(actions))

	for _, action := range actions {
		tasks = append(tasks, NewTask(action))
	}

	return tasks
}

// Invoke creates a new Task and runs it asynchronously.
func Invoke[T any](ctx context.Context, action Work[T]) Task[T] {
	return NewTask(action).Run(ctx)
}

// InvokeInSilence creates a new SilentTask and runs it asynchronously.
func InvokeInSilence(ctx context.Context, action SilentWork) SilentTask {
	return Invoke(
		ctx, func(taskCtx context.Context) (struct{}, error) {
			return struct{}{}, action(taskCtx)
		},
	)
}

// ContinueWith proceeds with the next task once the current one is finished.
func ContinueWith[T any, S any](ctx context.Context, currentTask Task[T], nextAction func(context.Context, T, error) (S, error)) Task[S] {
	return Invoke(
		ctx, func(taskCtx context.Context) (S, error) {
			result, err := currentTask.Outcome()

			return nextAction(taskCtx, result, err)
		},
	)
}

// ContinueWithNoResult proceeds with the next task once the current one is finished.
func ContinueWithNoResult[T any](ctx context.Context, currentTask Task[T], nextAction func(context.Context, T, error) error) SilentTask {
	return Invoke(
		ctx, func(taskCtx context.Context) (struct{}, error) {
			result, err := currentTask.Outcome()

			return struct{}{}, nextAction(taskCtx, result, err)
		},
	)
}

// ContinueInSilence proceeds with the next task once the current one is finished.
func ContinueInSilence(ctx context.Context, currentTask SilentTask, nextAction func(context.Context, error) error) SilentTask {
	return Invoke(
		ctx, func(taskCtx context.Context) (struct{}, error) {
			currentTask.Wait()

			return struct{}{}, nextAction(taskCtx, currentTask.Error())
		},
	)
}

// ContinueWithResult proceeds with the next task once the current one is finished.
func ContinueWithResult[T any](ctx context.Context, currentTask SilentTask, nextAction func(context.Context, error) (T, error)) Task[T] {
	return Invoke(
		ctx, func(taskCtx context.Context) (T, error) {
			currentTask.Wait()

			return nextAction(taskCtx, currentTask.Error())
		},
	)
}

func (t *task[T]) Outcome() (T, error) {
	<-t.done
	return t.outcome.result, t.outcome.err
}

func (t *task[T]) Error() error {
	<-t.done
	return t.outcome.err
}

func (t *task[T]) Wait() {
	<-t.done
}

func (t *task[T]) State() State {
	v := atomic.LoadInt32(&t.state)
	return State(v)
}

func (t *task[T]) Duration() time.Duration {
	return t.duration
}

func (t *task[T]) Run(ctx context.Context) Task[T] {
	go t.doRun(ctx)
	return t
}

func (t *task[T]) Execute(ctx context.Context) SilentTask {
	go t.doRun(ctx)
	return t
}

func (t *task[T]) Cancel() {
	// If the task was created but never started, transition directly to cancelled state
	// and close the done channel and set the error.
	if t.changeState(IsCreated, IsCancelled) {
		t.outcome = outcome[T]{err: errCancelled}
		close(t.done)
		return
	}

	// Attempt to cancel the task if it's in the running state
	if t.cancel != nil {
		select {
		case <-t.cancel:
			return
		default:
			close(t.cancel)
		}
	}
}

func (t *task[T]) doRun(ctx context.Context) {
	if !t.changeState(IsCreated, IsRunning) {
		return // Prevent from running the same task twice
	}

	// Notify everyone of the completion/error state
	defer close(t.done)

	// Execute the task
	startedAt := now().UnixNano()
	outcomeCh := make(chan outcome[T], 1)
	go func() {
		// Convert panics into standard errors for clients to handle gracefully
		defer func() {
			if r := recover(); r != nil {
				outcomeCh <- outcome[T]{err: fmt.Errorf("panic executing async task: %v \n %s", r, debug.Stack())}
			}
		}()

		r, e := t.action(ctx)
		outcomeCh <- outcome[T]{result: r, err: e}
	}()

	select {
	// In case of a manual task cancellation, set the outcome and transition
	// to the cancelled state.
	case <-t.cancel:
		t.duration = time.Nanosecond * time.Duration(now().UnixNano()-startedAt)
		t.outcome = outcome[T]{err: errCancelled}
		t.changeState(IsRunning, IsCancelled)
		return

	// In case of the context timeout or other error, change the state of the
	// task to cancelled and return right away.
	case <-ctx.Done():
		t.duration = time.Nanosecond * time.Duration(now().UnixNano()-startedAt)
		t.outcome = outcome[T]{err: ctx.Err()}
		t.changeState(IsRunning, IsCancelled)
		return

	// In case where we got an outcome (happy path).
	case o := <-outcomeCh:
		t.duration = time.Nanosecond * time.Duration(now().UnixNano()-startedAt)
		t.outcome = o
		t.changeState(IsRunning, IsCompleted)
		return
	}
}

func (t *task[T]) changeState(from, to State) bool {
	return atomic.CompareAndSwapInt32(&t.state, int32(from), int32(to))
}
