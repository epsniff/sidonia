package lifecycle

import (
	"sync"
)

type State int

const (
	// These are the possible life cycle states that a service may be in.

	// The service is freshly created and has not yet finished initializing.
	STATE_NEW State = 0
	// The service has finished initializing and is operating normally.
	STATE_RUNNING State = 5
	// The service is shutting down and may not be accepting any new requests.
	STATE_STOPPING State = 10
	// The service has finished shutting down and is completely unavailable.
	STATE_STOPPED State = 15
)

// A state machine for services which have lifecycles where they initialize, run, and stop.
// Provides the ability to wait for lifecycle events, such as "service has finished 
// initializing" or "service has finished shutting down."
type LifeCycle struct {
	mutex   *sync.Mutex
	onEntry map[State](chan interface{})
	state   State
}

func NewLifeCycle() *LifeCycle {
	return &LifeCycle{
		mutex:   &sync.Mutex{},
		onEntry: make(map[State]chan interface{}),
	}
}

// Wait until the service enters the given state. If the service enters another state that
// makes it impossible that the requested state will ever be entered, this function will also
// return in that case. For example, if you're waiting for STATE_RUNNING but there's an error
// during initialization and the service jumps to STATE_STOPPED, then this function will return
// STATE_STOPPED instead of blocking forever.
func (this *LifeCycle) WaitForState(state State) State {
	waitChan := this.waitChanUntilStateEntered(state)
	item := <-waitChan // Loop the value back into the chan in case someone else is waiting
	waitChan <- item

	return this.GetState()

}

// Internal: return a channel that will return a value once the desired state (or any later
// state) is entered.
func (this *LifeCycle) waitChanUntilStateEntered(state State) chan interface{} {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.state >= state {
		// We're already in the requested state, or some later state. Return a channel that 
		// will trigger immediately.
		immediateChan := make(chan interface{}, 1)
		immediateChan <- true
		return immediateChan
	}

	ch, alreadyExists := this.onEntry[state]
	if alreadyExists {
		return ch
	}
	ch = make(chan interface{}, 1)
	this.onEntry[state] = ch
	return ch
}

// Change to the given state in this life cycle. For example, when the service is done
// initializing you should probably change it's state to STATE_RUNNING. This includes
// unblocking any goroutines that might be waiting for the state transition.
//
// Life cycle states must happen in order. You can skip states, but never go backward. For 
// example, a service that is RUNNING may never go back to NEW, and a service that is
// STOPPED may never go back to RUNNING.
func (this *LifeCycle) Transition(newState State) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if newState <= this.state {
		panic("Can't transition back to a previous state")
	}

	this.transition_holdinglock(newState)
}

// Get the current state in the life cycle (for example, "still initializing").
func (this *LifeCycle) GetState() State {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	return this.state
}

func (this *LifeCycle) transition_holdinglock(newState State) {
	this.state = newState

	for notifyState, notifyChan := range this.onEntry {
		if notifyState <= newState {
			notifyChan <- true
			delete(this.onEntry, newState)
		}
	}
}
