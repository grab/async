# Async

## What is package async
Package async simplifies the implementation of orchestration patterns for concurrent systems. It is similar to Java Future or JS Promise, which makes life much easier when dealing with asynchronous operation and concurrent processing. Golang is excellent in term of parallel programming. However, dealing with goroutine and channels could be a big headache when business logic gets complicated. Wrapping them into higher-level functions brings code much better readability and developers a ease of thinking.
                                                                                              
Currently, this packageg includes:

* Asynchronous tasks with cancellations, context propagation and state.
* Task chaining by using continuations.
* Fork/join pattern - running a bunch of work and waiting for everything to finish.
* Throttling pattern - throttling task execution on a specified rate.
* Spread pattern - spreading tasks across time.
* Partition pattern - partitioning data concurrently.
* Repeat pattern - repeating a certain task at a specified interval.
* Batch pattern - batching many tasks into a single one with individual continuations.

## Concept
**Task** is a basic concept like Future in Java. You can create a Task with an executable function which takes in context and returns result and error.
```
task := NewTask(func(context.Context) (interface{}, error) {
    // run the job
   return res, err
})
```
#### Get the result
The function will be evaluated asynchronously. You can query whether it's completed by calling task.State(), which would be a non-blocking function. Alternative, you can wait for the response with task.Outcome(), which will block the execution until the job is done. These 2 functions are quite similar to Future.isDone() or Future.get()

#### Cancelling
There could be case that we don't care about the result anymore some time after execution. In this case, the task can be aborted by invoking task.Cancel().

#### Chaining
To have a follow-up action after the task, we can simply call ContinueWith(). This could be very useful to create a chain of processing, or like have a teardown process after the job.

## Examples
For example, if want to upload numerous files efficiently. There are multiple strategies you can take 
Given file uploading function like:
```
func upload(context.Context) (interface{}, error){
    // do file uploading 
    return res, err
}

```
 
#### Fork join
The main characteristic for Fork join task is to spawn new subtasks running concurrently. They could be different parts of the main task which can be running independently.  The following code example illustrates how you can send files to S3 concurrently with few lines of code.


```
func uploadFilesConcurrently(files []string) {
	tasks := []Tasks{}
		for _, file := files {
		tasks = append(tasks, NewTask(upload(file)))
	}
   ForkJoin(context.Background(), tasks)
}
```

#### Invoke All
The Fork Join may not apply to every cases imagining the number of tasks go crazy. In that case, the number of concurrently running tasks, goroutines and CPU utilisation would overwhelm the node. One solution is to constraint the maximum concurrency. InvokeAll is introduced for this purpose, it's like maintaining a fixed size of goroutine pool which attempt serve the given tasks with shortest time.
```
InvokeAll(context.Background(), concurrency, tasks)
```

#### Spread
Sometimes we don't really care about the concurrency but just want to make sure the tasks could be finished with certain time period. Spread function would be useful in this case by spreading the tasks evenly in given period.
```
Spread(context.Background(), period, tasks)
```
For example, if we want to send 50 files within 10 seconds, the Spread function would start to run the task every 0.2 second. An assumption made here is that every task takes similar period of time. To have more sophisticated model, we may need to have adaptive learning model to derive the task duration from characteristics or parameters of distinct tasks.

