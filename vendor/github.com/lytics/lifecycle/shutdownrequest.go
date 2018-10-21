package lifecycle

import (
	"sync/atomic"
)

/*
ShutdownRequest solves the problem of quickly and cleanly shutting down a goroutine. When a 
goroutine is reading incoming work items from a channel and executing them, using the quit 
channel approach may not immediately shut down the worker:

	keepLooping := true
	for keepLooping {
		select {
    		workItem := <-workChan:
		    	doWork(workItem)
	    	<-quitChan:
		    	keepLooping = false
		}
	}

In the above example, the worker may continue processing workChan for many iterations until it
happens to read from quitChan. Using ShutdownRequest you can guarantee termination within one
iteration:

	for !shutdownRequest.IsShutdownRequested() {
		select {
		case toAdd := <-onesChan:
			atomic.AddInt32(&sum, int32(toAdd))
		case <-shutdownRequest.GetShutdownRequestChan():
			break
		}
	}
*/
type ShutdownRequest struct {
	shutdownRequestChan chan interface{}
	isShutdownRequested int32
}

func NewShutdownRequest() *ShutdownRequest {
	return &ShutdownRequest{
		shutdownRequestChan: make(chan interface{}),
	}
}

// Call this in the server main loop to check whether it should shut down. Use this together
// with the shutdown request channel.
func (this *ShutdownRequest) IsShutdownRequested() bool {
	// If the channel is still open, shutdown has not yet been requested
	select {
	case <-this.shutdownRequestChan:
		// Since the channel never receives any messages, this code will execute only if the 
		// channel was closed.
		return true
	default:
		// Make the receive non-blocking, so if the channel is open, we'll fall through.
	}
	return false // channel is still open
}

// Get the channel that will be closed when shutdown is requested. Use the returned
// channel in the server's main loop select statement. Use the returned channel in a select 
// statement, all inside a for loop whose predicate checks IsShutdownRequested() (see above).
func (this *ShutdownRequest) GetShutdownRequestChan() chan interface{} {
	return this.shutdownRequestChan
}

// Call this function to asynchronously indicate that the owning service should shut down.
func (this *ShutdownRequest) RequestShutdown() {
	// The compare-and-swap operation makes sure we only try to close the channel once.
	// Attempting to close it multiple times will cause a runtime panic.
	if atomic.CompareAndSwapInt32(&this.isShutdownRequested, 0, 1) {
		close(this.shutdownRequestChan)
	}
}
