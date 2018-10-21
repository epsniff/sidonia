lifecycle
=========

Lifecycle is a go package that helps you manage goroutines that provide services to other goroutines.

Things you can do:
 - Have one goroutine wait for another goroutine to initalize itself and be ready to work. The server goroutine calls ```LifeCycle.Transition(STATE_RUNNING)``` and the client calls ```LifeCycle.WaitForState(STATE_RUNNING)```
 - Wait for a goroutine service to finish shutting down. This is handy for testing graceful shutdown, and it's friendlier than using WaitGroups manually.

Separately, a ShutdownRequest is a good way to quickly shut down the main loop of a server without an unbounded of iterations that can occur with the "quit channel" approach:

```go
for !shutdownRequest.IsShutdownRequested() {
    select {
        case toAdd := <-onesChan:
            atomic.AddInt32(&sum, int32(toAdd))
        case <-shutdownRequest.GetShutdownRequestChan():
            break
        }
}
```

You might be wondering what is wrong with the "quit channel" approach:
```go
// This is the wrong way to do it, for educational purposes only:
keepLooping := true
for keepLooping {
    select {
        workItem := <-workChan:
            doWork(workItem)
        <-quitChan:
            keepLooping = false
    }
}
```

In the "quit channel" approach, if the workChan channel is readable, the for loop might execute a few more times (actually unbounded) before the quitChan channel is read. In Go, if a select statement has multiple channels that are ready, one of them will be pseudorandomly picked. This implies that you need an exit flag outside of the select statement, which is what ShutdownRequest provides.
