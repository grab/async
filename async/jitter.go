// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import (
	"context"
	"math/rand"
	"time"
)

// DoJitter adds a random jitter before executing doFn, then returns the jitter duration.
// Why jitter?
//
// http://highscalability.com/blog/2012/4/17/youtube-strategy-adding-jitter-isnt-a-bug.html
func DoJitter(doFn func(), maxJitterDurationInMilliseconds int) int {
	randomJitterDuration := waitForRandomJitter(maxJitterDurationInMilliseconds)

	doFn()

	return randomJitterDuration
}

// AddJitterT adds a random jitter before executing the given Task. Why jitter?
//
// http://highscalability.com/blog/2012/4/17/youtube-strategy-adding-jitter-isnt-a-bug.html
func AddJitterT[T any](t Task[T], maxJitterDurationInMilliseconds int) Task[T] {
	return NewTask(
		func(ctx context.Context) (T, error) {
			waitForRandomJitter(maxJitterDurationInMilliseconds)

			t.Execute(ctx)

			return t.Outcome()
		},
	)
}

// AddJitterST adds a random jitter before executing the given SilentTask. Why jitter?
//
// http://highscalability.com/blog/2012/4/17/youtube-strategy-adding-jitter-isnt-a-bug.html
func AddJitterST(t SilentTask, maxJitterDurationInMilliseconds int) SilentTask {
	return NewSilentTask(
		func(ctx context.Context) error {
			waitForRandomJitter(maxJitterDurationInMilliseconds)

			t.Execute(ctx)

			return t.Error()
		},
	)
}

func waitForRandomJitter(maxJitterDurationInMilliseconds int) int {
	rand.Seed(time.Now().UnixNano())
	min := 0
	max := maxJitterDurationInMilliseconds

	randomJitterDuration := rand.Intn(max-min+1) + min

	<-time.After(time.Duration(randomJitterDuration) * time.Millisecond)

	return randomJitterDuration
}
