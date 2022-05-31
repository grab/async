package async

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gammazero/deque"
)

const (
	// The default maximum duration that a worker can stay idle before one of them gets killed.
	defaultIdleTimeout = 2 * time.Second
)

// queuedTask represents a task to be executed using the provided context.
type queuedTask struct {
	task SilentTask
	ctx  context.Context
}

// run executes the queued task using the provided context.
func (qt *queuedTask) run() {
	qt.task.Execute(qt.ctx)
	// Worker is already running on an independent goroutine. We must wait for
	// the task to complete before releasing the worker for the next task.
	qt.task.Wait()
}

// cancel cancels the queued task.
func (qt *queuedTask) cancel() {
	qt.task.Cancel()
}

// WorkerPool is similar to a thread pool in Java, where the number of concurrent
// goroutines processing requests does not exceed the configured maximum.
type WorkerPool struct {
	*workerPoolConfigs

	taskQueue          chan *queuedTask
	workerQueue        chan *queuedTask
	stoppedChan        chan struct{}
	stopSignal         chan struct{}
	waitingQueue       deque.Deque
	workerCount        int
	stopLock           sync.Mutex
	stopOnce           sync.Once
	stopped            bool
	pendingSize        int32
	waitBeforeShutdown bool
}

// NewWorkerPool creates and starts a pool of worker goroutines.
//
// `maxSize` specifies the maximum number of workers that can execute tasks
// concurrently. When there's no incoming tasks, workers get killed 1-by-1
// until there's no remaining workers.
func NewWorkerPool(options ...WorkerPoolOption) *WorkerPool {
	pool := &WorkerPool{
		workerPoolConfigs: &workerPoolConfigs{
			maxSize:     runtime.NumCPU(),
			idleTimeout: defaultIdleTimeout,
		},
		taskQueue:   make(chan *queuedTask, 1),
		workerQueue: make(chan *queuedTask),
		stopSignal:  make(chan struct{}),
		stoppedChan: make(chan struct{}),
	}

	for _, o := range options {
		o(pool.workerPoolConfigs)
	}

	// Start the task dispatcher.
	go pool.dispatch()

	return pool
}

// Size returns the maximum number of concurrent workers.
func (p *WorkerPool) Size() int {
	return p.maxSize
}

// Stop stops this worker pool and waits for the running tasks to complete. Pending tasks
// that are still in the queue will get cancelled. New tasks must not be submitted to the
// worker pool after calling Stop().
//
// Note: to avoid memory leak, clients MUST always call Stop() or StopWait() when this
// worker pool is no longer needed.
func (p *WorkerPool) Stop() {
	p.stop(false)
}

// StopWait stops this worker pool and waits for the running tasks + all queued tasks to
// complete. New tasks must not be submitted to the worker pool after calling Stop().
//
// Note: to avoid memory leak, clients MUST always call Stop() or StopWait() when this
// worker pool is no longer needed.
func (p *WorkerPool) StopWait() {
	p.stop(true)
}

// Stopped returns true if this worker pool has been stopped.
func (p *WorkerPool) Stopped() bool {
	p.stopLock.Lock()
	defer p.stopLock.Unlock()

	return p.stopped
}

// Submit enqueues a function for a worker to execute.
//
// Submit will not block regardless of the number of tasks submitted. Each task is given
// to an available worker or to a newly started worker. If there's no available workers,
// and no new workers can be created due to the configured maximum, then the task is put
// onto a waiting queue.
//
// When the waiting queue is not empty, incoming tasks will go into the queue immediately.
// Tasks are removed from the waiting queue as workers become available.
//
// As long as no new tasks arrive, one idle worker will get killed periodically until no
// more workers are left. Since starting new goroutines is cheap & quick, there's no need
// to retain idle workers indefinitely.
func (p *WorkerPool) Submit(ctx context.Context, task SilentTask) {
	if task != nil {
		p.taskQueue <- &queuedTask{
			task: task,
			ctx:  ctx,
		}
	}
}

// WaitingQueueSize returns the count of tasks in the waiting queue.
func (p *WorkerPool) WaitingQueueSize() int {
	return int(atomic.LoadInt32(&p.pendingSize))
}

// Pause causes all workers to wait on the given context, thereby making them unavailable
// to run tasks. Pause returns when all workers are waiting. Tasks can still be queued but
// won't get executed until `ctx` is cancelled or times out.
//
// Calling Pause when the worker pool is already paused causes Pause to wait until previous
// pauses are cancelled. This allows a goroutine to take control of pausing and un-pausing
// the pool as soon as other goroutines have un-paused it.
//
// When this worker pool is stopped, workers are un-paused and queued tasks may be executed
// during StopWait.
func (p *WorkerPool) Pause(ctx context.Context) {
	p.stopLock.Lock()
	defer p.stopLock.Unlock()

	if p.stopped {
		return
	}

	ready := new(sync.WaitGroup)
	ready.Add(p.maxSize)

	for i := 0; i < p.maxSize; i++ {
		pauseTask := NewSilentTask(
			func(taskCtx context.Context) error {
				ready.Done()

				select {
				case <-taskCtx.Done():
				case <-p.stopSignal:
				}

				return nil
			},
		)

		p.Submit(ctx, pauseTask)
	}

	// Wait for workers to all be paused
	ready.Wait()
}

// dispatch sends the next queued task to an available worker.
func (p *WorkerPool) dispatch() {
	defer close(p.stoppedChan)

	var idle bool
	var wg sync.WaitGroup

Loop:
	for {
		// As long as the waiting queue is not empty, incoming tasks will go directly
		// into the waiting queue. We'll try to run all tasks from this queue first.
		// Once the waiting queue is empty, then go back to submitting incoming tasks
		// directly to available workers.
		if p.waitingQueue.Len() != 0 {
			if !p.processWaitingQueue(&wg) {
				break Loop
			}

			continue
		}

		select {
		case task, ok := <-p.taskQueue:
			if !ok {
				break Loop
			}

			// Got a task to do.
			select {
			case p.workerQueue <- task:
			default:
				// Create a new worker, if not at max.
				if p.workerCount < p.maxSize {
					wg.Add(1)
					go func() {
						task.run()
						p.spawnWorker(&wg)
					}()
					p.workerCount++

				} else {
					// Enqueue task to be executed by next available worker.
					p.pushBack(task)

				}
			}

			idle = false

		case <-time.After(p.idleTimeout):
			// Timed out waiting for a new task to arrive. Kill an available worker if
			// this worker pool has been idle for the whole duration.
			if idle && p.workerCount > 0 {
				p.killIdleWorker()
			}

			idle = true
		}
	}

	if p.waitBeforeShutdown {
		p.drainQueuedTasks(false)
	} else {
		p.drainQueuedTasks(true)
	}

	// Stop all workers as they become available.
	for p.workerCount > 0 {
		p.workerQueue <- nil
		p.workerCount--
	}

	wg.Wait()
}

// spawnWorker creates a new worker to execute tasks and stops it on receiving a nil task.
func (p *WorkerPool) spawnWorker(wg *sync.WaitGroup) {
	defer wg.Done()

	for task := range p.workerQueue {
		if task == nil {
			return
		}

		task.run()
	}
}

// stop tells the dispatcher to exit, and whether to complete the queued tasks.
func (p *WorkerPool) stop(waitBeforeShutdown bool) {
	p.stopOnce.Do(
		func() {
			// Indicate that worker pool is stopping, to unpause all paused workers.
			close(p.stopSignal)
			// Acquire stopLock to wait for any pause in progress to complete.
			p.stopLock.Lock()
			// The stopped flag prevents this worker pool from being paused again. This
			// makes it safe to close the taskQueue.
			p.stopped = true
			p.stopLock.Unlock()
			p.waitBeforeShutdown = waitBeforeShutdown
			// Close task queue and wait for currently running tasks to finish.
			close(p.taskQueue)
		},
	)

	<-p.stoppedChan
}

// processWaitingQueue puts new tasks onto the waiting queue, and removes tasks from the
// waiting queue as workers become available. If the queue's length reaches the configured
// threshold, a configured amount of new workers will be spawned to increase throughput.
// These new workers will eventually get killed once they stay idle at a later time. Returns
// false if this worker pool was stopped.
func (p *WorkerPool) processWaitingQueue(wg *sync.WaitGroup) bool {
	select {
	case t, ok := <-p.taskQueue:
		if !ok {
			return false
		}

		curQueueLength := p.pushBack(t)
		if curQueueLength == int32(p.burstQueueThreshold) {
			for i := 0; i < p.burstCapacity; i++ {
				wg.Add(1)
				go p.spawnWorker(wg)
				p.workerCount++
			}
		}

	case p.workerQueue <- p.peekFront():
		// A task was given to an available worker.
		p.popFront()

	}

	return true
}

func (p *WorkerPool) killIdleWorker() bool {
	select {
	case p.workerQueue <- nil:
		// Sent kill signal to worker.
		p.workerCount--
		return true
	default:
		// No ready workers. All, if any, workers are busy.
		return false
	}
}

// drainQueuedTasks continuously pops a task from the waiting queue and either cancel
// or gives it to an available worker until the queue is empty.
func (p *WorkerPool) drainQueuedTasks(toCancel bool) {
	for p.waitingQueue.Len() != 0 {
		t := p.popFront()

		if toCancel {
			t.cancel()
			continue
		}

		// Give a task to the next available worker
		p.workerQueue <- t
	}
}

// pushBack pushes a task to the back of the queue
func (p *WorkerPool) pushBack(task *queuedTask) int32 {
	p.waitingQueue.PushBack(task)

	queueLength := int32(p.waitingQueue.Len())
	atomic.StoreInt32(&p.pendingSize, queueLength)

	return queueLength
}

// popFront removes and returns the task at the front of the queue
func (p *WorkerPool) popFront() *queuedTask {
	t := p.waitingQueue.PopFront().(*queuedTask)
	atomic.StoreInt32(&p.pendingSize, int32(p.waitingQueue.Len()))

	return t
}

// peekFront returns but not removes the task at the front of the queue
func (p *WorkerPool) peekFront() *queuedTask {
	return p.waitingQueue.Front().(*queuedTask)
}
