package async

import (
	"runtime"
	"time"
)

type workerPoolConfigs struct {
	maxSize             int
	idleTimeout         time.Duration
	burstQueueThreshold int
	burstCapacity       int
}

type WorkerPoolOption func(*workerPoolConfigs)

// WithMaxSize sets the maximum size of the worker pool under normal condition.
func WithMaxSize(maxSize int) WorkerPoolOption {
	return func(configs *workerPoolConfigs) {
		if maxSize <= 0 {
			configs.maxSize = runtime.NumCPU()
			return
		}

		configs.maxSize = maxSize
	}
}

// WithIdleTimeout sets the maximum duration that a worker can stay idle before one of them gets killed.
func WithIdleTimeout(idleTimeout time.Duration) WorkerPoolOption {
	return func(configs *workerPoolConfigs) {
		configs.idleTimeout = idleTimeout
	}
}

// WithBurst sets the threshold for the waiting queue at which point the maximum size
// of the worker pool will be increased by the given capacity.
func WithBurst(burstQueueThreshold int, burstCapacity int) WorkerPoolOption {
	return func(configs *workerPoolConfigs) {
		configs.burstQueueThreshold = burstQueueThreshold
		configs.burstCapacity = burstCapacity
	}
}
