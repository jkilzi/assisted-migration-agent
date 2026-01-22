// Package scheduler implements a worker pool for executing async work with futures.
//
// The scheduler manages a fixed pool of workers that execute work functions
// concurrently. Work is submitted via AddWork and returns a Future that can
// be used to retrieve the result or cancel the work.
//
// # Architecture Overview
//
//	┌─────────────────────────────────────────────────────────────────────┐
//	│                           Scheduler                                 │
//	│                                                                     │
//	│  ┌──────────────┐      ┌──────────────┐      ┌──────────────┐       │
//	│  │   Worker 1   │      │   Worker 2   │      │   Worker N   │       │
//	│  └──────────────┘      └──────────────┘      └──────────────┘       │
//	│         ▲                     ▲                     ▲               │
//	│         │                     │                     │               │
//	│         └─────────────────────┼─────────────────────┘               │
//	│                               │                                     │
//	│                        ┌──────┴──────┐                              │
//	│                        │  dispatch() │                              │
//	│                        └──────┬──────┘                              │
//	│                               │                                     │
//	│  ┌────────────────────────────┴────────────────────────────┐        │
//	│  │                      Work Queue                         │        │
//	│  │  [work1] [work2] [work3] ...                            │        │
//	│  └─────────────────────────────────────────────────────────┘        │
//	│                               ▲                                     │
//	│                               │                                     │
//	│                        AddWork(fn)                                  │
//	└─────────────────────────────────────────────────────────────────────┘
//
// # Core Components
//
// Scheduler:
//   - Manages a pool of N workers (configured at creation)
//   - Maintains a work queue for pending work requests
//   - Runs an event loop dispatching work to available workers
//   - Supports graceful shutdown via Close()
//
// Worker:
//   - Executes a single work function
//   - Returns to the worker pool after completion
//   - Recovers from panics and reports them as errors
//
// Future:
//   - Represents a pending result from submitted work
//   - Provides a channel to receive the result
//   - Supports cancellation via Stop()
//
// # Work Execution Flow
//
//  1. Client calls AddWork(fn)
//     │
//     ▼
//  2. Scheduler creates workRequest with:
//     - Work function (fn)
//     - Result channel (buffered, size 1)
//     - Cancellable context (derived from main context)
//     │
//     ▼
//  3. workRequest sent to scheduler's work channel
//     │
//     ▼
//  4. Scheduler's run() loop receives work:
//     - Pushes to workQueue
//     - Calls dispatch()
//     │
//     ▼
//  5. dispatch() pairs available workers with pending work:
//     - While workers > 0 AND workQueue > 0:
//     - Pop work from queue
//     - Pop worker from pool
//     - Launch goroutine: worker.Work(request)
//     │
//     ▼
//  6. Worker executes work function:
//     - Calls fn(ctx) with cancellable context
//     - Sends Result{Data, Err} to result channel
//     - Signals completion via done channel
//     - Returns to worker pool
//     │
//     ▼
//  7. Client receives result via future.C() channel
//
// # Future Mechanism
//
// When AddWork is called, it returns a Future immediately. The Future provides:
//
//   - C() chan Result: Channel that will receive exactly one result when work completes
//   - Stop(): Cancels the work's context (signals cancellation to the work function)
//
// Usage pattern:
//
//	future := scheduler.AddWork(func(ctx context.Context) (any, error) {
//	    // Do work, check ctx.Done() for cancellation
//	    return result, nil
//	})
//
//	// Option 1: Block until complete
//	result := <-future.C()
//	if result.Err != nil {
//	    // Handle error
//	}
//	data := result.Data
//
//	// Option 2: Select with timeout or cancellation
//	select {
//	case result := <-future.C():
//	    // Handle result
//	case <-ctx.Done():
//	    future.Stop()  // Cancel the work
//	}
//
// # Worker Pool Mechanism
//
// The scheduler maintains two queues:
//
//	┌─────────────────────────────────────────────────────────────────────┐
//	│  Workers Queue (available workers)                                  │
//	│  [worker1] [worker2] [worker3] ...                                  │
//	│                                                                     │
//	│  - Initially populated with N workers                               │
//	│  - Workers removed when assigned work                               │
//	│  - Workers returned after completing work                           │
//	└─────────────────────────────────────────────────────────────────────┘
//
//	┌─────────────────────────────────────────────────────────────────────┐
//	│  Work Queue (pending work requests)                                 │
//	│  [request1] [request2] [request3] ...                               │
//	│                                                                     │
//	│  - Work added when no workers immediately available                 │
//	│  - Drained by dispatch() when workers become available              │
//	└─────────────────────────────────────────────────────────────────────┘
//
// Worker Lifecycle:
//
//	┌───────────┐     dispatch()      ┌───────────┐
//	│  Idle     │ ─────────────────► │  Working  │
//	│ (in pool) │                     │           │
//	└───────────┘                     └─────┬─────┘
//	      ▲                                 │
//	      │         done channel            │
//	      └─────────────────────────────────┘
//
// # Event Loop (run method)
//
// The scheduler runs an event loop handling three events:
//
//	for {
//	    select {
//	    case w := <-s.work:       // New work submitted
//	        s.workQueue.Push(w)
//	        s.dispatch()
//
//	    case <-s.done:            // Worker completed
//	        s.workers.Push(newWorker(...))
//	        s.dispatch()          // Try to assign queued work
//
//	    case <-s.close:           // Shutdown requested
//	        s.wg.Wait()           // Wait for in-flight work
//	        return
//	    }
//	}
//
// The dispatch() function is called both when new work arrives AND when
// workers complete, ensuring work is assigned as soon as workers are available.
//
// # Panic Recovery
//
// Workers recover from panics in work functions:
//
//	defer func() {
//	    if rec := recover(); rec != nil {
//	        r.c <- Result{Err: fmt.Errorf("worker panicked: %v", rec)}
//	    }
//	}()
//
// This ensures:
//   - Panics don't crash the scheduler
//   - The future receives an error result
//   - The worker returns to the pool
//
// # Cancellation
//
// Each work request gets a context derived from the scheduler's main context:
//
//	ctx, cancel := context.WithCancel(s.mainCtx)
//
// Cancellation hierarchy:
//   - future.Stop() → Cancels individual work's context
//   - scheduler.Close() → Cancels main context (all work)
//
// Work functions should check ctx.Done() to respond to cancellation:
//
//	func(ctx context.Context) (any, error) {
//	    for {
//	        select {
//	        case <-ctx.Done():
//	            return nil, ctx.Err()
//	        default:
//	            // Continue work
//	        }
//	    }
//	}
//
// # Graceful Shutdown
//
// Close() performs graceful shutdown:
//
//  1. Cancels main context (signals all work to stop)
//  2. Sends close signal to event loop
//  3. Event loop waits for all in-flight workers (wg.Wait())
//  4. Event loop exits and closes done channel
//  5. Close() returns after receiving from done channel
//
// Close() is idempotent (uses sync.Once).
//
// # Usage Example
//
//	// Create scheduler with 4 workers
//	sched := scheduler.NewScheduler(4)
//	defer sched.Close()
//
//	// Submit work
//	future := sched.AddWork(func(ctx context.Context) (any, error) {
//	    // Simulate work
//	    time.Sleep(100 * time.Millisecond)
//	    return "done", nil
//	})
//
//	// Wait for result
//	result := <-future.C()
//	if result.Err != nil {
//	    log.Printf("Work failed: %v", result.Err)
//	} else {
//	    log.Printf("Work completed: %v", result.Data)
//	}
package scheduler
