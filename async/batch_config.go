// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import "time"

type batcherConfigs struct {
	autoProcessSize       int
	autoProcessInterval   time.Duration
	shutdownGraceDuration time.Duration
}

type BatcherOption func(*batcherConfigs)

// WithAutoProcessSize sets the limit at which Batcher will automatically process
// the pending tasks. If `autoProcessSize <= 0`, the default behavior applies: no
// auto processing will be done based on size.
func WithAutoProcessSize(autoProcessSize int) BatcherOption {
	return func(configs *batcherConfigs) {
		configs.autoProcessSize = autoProcessSize
	}
}

// WithAutoProcessInterval sets the interval at which Batcher will automatically process
// the pending tasks. If `autoProcessDurationInMilliseconds <= 0`, the default behavior
// applies: no periodic auto processing will be done.
//
// Note: if periodic auto processing is ON, clients MUST call Batcher.Shutdown() to clean
// up the timer goroutine properly in order to avoid memory leak.
func WithAutoProcessInterval(autoProcessIntervalInMilliseconds time.Duration) BatcherOption {
	return func(configs *batcherConfigs) {
		configs.autoProcessInterval = autoProcessIntervalInMilliseconds
	}
}

// WithShutdownGraceDuration specifies how long Batcher will wait for the Shutdown operation
// to complete before returning. If `shutdownDurationInMilliseconds <= 0`, Batcher will block
// and wait until the shutdown operation fully completes.
func WithShutdownGraceDuration(shutdownDurationInMilliseconds time.Duration) BatcherOption {
	return func(configs *batcherConfigs) {
		configs.shutdownGraceDuration = shutdownDurationInMilliseconds
	}
}
