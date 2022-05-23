# Async

## Why you should consider this package
Package `async` simplifies the implementation of orchestration patterns for concurrent systems. It is similar to 
Java Future or JS Promise, which makes life much easier when dealing with asynchronous operation and concurrent 
processing. Golang is excellent in terms of parallel programming. However, dealing with Go routines and channels 
could be a big headache when business logic gets complicated. Wrapping them into higher-level functions improves
code readability significantly and makes it easier for engineers to reason about the system's behaviours.
                                                                                              
Currently, this package includes:

* Asynchronous tasks with cancellations, context propagation and state.
* Task chaining by using continuations.
* Fork/join pattern - running a batch of tasks in parallel and blocking until all finish.
* Concurrency cap - running a batch of tasks concurrently with a cap on max concurrency level.
* Throttling pattern - throttling task execution at a specified rate.
* Spread pattern - spreading tasks within a specified duration.
* Repeat pattern - repeating a task on a pre-determined interval.
* Batch pattern - batching many tasks to be processed together with individual continuations.
* Partition pattern - dividing data into partitions concurrently.

## Concept
**Task** is a basic concept like `Future` in Java. You can create a `Task` using an executable function which takes 
in `context.Context`, then returns error and an optional result.

```go
task := NewTask(func(context.Context) (animal, error) {
    // run the job
    return res, err
})

silentTask := NewSilentTask(func(context.Context) error {
    // run the job
    return err
})
```

### Get the result
The function will be executed asynchronously. You can query whether it's completed by calling `task.State()`, which 
is a non-blocking function. Alternative, you can wait for the response using `task.Outcome()` or `silentTask.Wait()`, 
which will block the execution until the task is done. These functions are quite similar to the equivalents in Java
`Future.isDone()` or `Future.get()`

### Cancelling
There could be case that we don't care about the result anymore some time after execution. In this case, a task can 
be aborted by invoking `task.Cancel()`.

### Chaining
To have a follow-up action after a task is done, you can use the provided family of `Continue` functions. This could 
be very useful to create a chain of processing, or to have a teardown process at the end of a task.

## Features
 
### Fork join
`ForkJoin` is meant for running multiple subtasks concurrently. They could be different parts of the main task which 
can be executed independently. The following code example illustrates how you can send files to S3 concurrently with 
a few lines of code.

```go
func uploadFilesConcurrently(files []string) {
    var tasks []Task[string]
    for _, file := range files {
        f := file
        
        tasks = append(tasks, NewTask(func(ctx context.Context) (string, error) {
            return upload(ctx, f)
        }))
    }

    ForkJoin(context.Background(), tasks)
}

func upload(ctx context.Context, file string) (string, error){
    // do file uploading
    return "", nil
}
```

### Concurrency cap
`ForkJoin` is not suitable when the number of tasks is huge. In this scenario, the number of concurrent Go routines
would likely overwhelm a node and consume too much CPU resources. One solution is to put a cap on the max concurrency
level. `RunWithConcurrencyLevelC` and `RunWithConcurrencyLevelS` were created for this purpose. Internally, it's like 
maintaining a fixed-size worker pool which aims to execute the given tasks as quickly as possible without violating 
the given constraint.

```go
// RunWithConcurrencyLevelC runs the given tasks up to the max concurrency level.
func RunWithConcurrencyLevelC[T SilentTask](ctx context.Context, concurrencyLevel int, tasks <-chan T) SilentTask

// RunWithConcurrencyLevelS runs the given tasks up to the max concurrency level.
func RunWithConcurrencyLevelS[T SilentTask](ctx context.Context, concurrencyLevel int, tasks []T) SilentTask
```

### Throttle
Sometimes you don't really care about the concurrency level but just want to execute the tasks at a particular rate.
The `Throttle` function would come in handy in this case.

```go
// Throttle runs the given tasks at the specified rate.
func Throttle[T SilentTask](ctx context.Context, tasks []T, rateLimit int, every time.Duration) SilentTask
```

For example, if you want to send 4 files every 2 seconds, the `Throttle` function will start a task every 0.5 second.

### Spread
Instead of starting all tasks at once with `ForkJoin`, you can also spread the starting points of our tasks evenly
within a certain duration using the `Spread` function.

```go
// Spread evenly spreads the tasks within the specified duration.
func Spread[T SilentTask](ctx context.Context, tasks []T, within time.Duration) SilentTask
```

For example, if you want to send 50 files within 10 seconds, the `Spread` function would start a task every 0.2s.

### Repeat

In cases where you need to repeat a background task on a pre-determined interval, `Repeat` is your friend. The 
returned `SilentTask` can then be used to cancel the repeating task at any time.

```go
// Repeat executes the given SilentWork asynchronously on a pre-determined interval.
func Repeat(ctx context.Context, interval time.Duration, action SilentWork) SilentTask
```

### Batch

Instead of executing a task immediately whenever you receive an input, sometimes, it might be more efficient to
create a batch of inputs and process all in one go.

See `batch_test.go` for a detailed example of how to use the `Batch` feature.

### Partition

When you receive a lot of data concurrently, it might be useful to divide the data into separate partitions before
consuming.  

See `partition_test.go` for a detailed example of how to use the `Partition` feature.