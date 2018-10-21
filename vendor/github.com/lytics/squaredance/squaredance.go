// Squaredance provides a simple interface for spawning stoppable child tasks
package squaredance

import (
	"fmt"
	"runtime"
	"sync"
)

// The Caller allows spawning, stopping and waiting on child tasks
type Caller interface {
	// Wait blocks until all applicable children close
	Wait() error
	// Stops all children
	Stop()
	// Spawns a child task
	Spawn(func(Follower) error)
}

// The follower is passed to the spawned function, allowing access to
// the stop channel, as well as spawning additional peers
type Follower interface {
	// Closed when the following task should stop
	StopChan() <-chan struct{}
	// Spawn a new (peer) child task
	Spawn(func(Follower) error)
}

func IsPanic(e error) bool {
	_, ok := e.(*PanicError)
	return ok
}

// PanicError is returned on child panic
type PanicError struct {
	// Value returned from recover()
	Panic interface{}
	Stack []byte
}

func (p *PanicError) Error() string {
	return fmt.Sprintf("Panic from child: %v\n%s", p.Panic, p.Stack)
}

// Create a new parent Caller
func NewCaller() Caller {
	return &caller{
		stopchan: make(chan struct{}),
	}
}

type caller struct {
	childwg  sync.WaitGroup
	mu       sync.Mutex
	stopchan chan struct{}
	firsterr error
	stopped  bool
}

func (t *caller) Spawn(sf func(Follower) error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.stopped {
		return
	}
	t.childwg.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				p := &PanicError{r, make([]byte, 100000)}
				wrote := runtime.Stack(p.Stack, false)
				if wrote < 100000 {
					p.Stack = p.Stack[:wrote]
				}
				t.setError(p)
			}
			t.childwg.Done()
		}()
		err := sf(t)
		if err != nil {
			t.setError(err)
		}
	}()
}

func (t *caller) Wait() error {
	t.childwg.Wait()
	return t.firsterr
}

func (t *caller) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.stopped {
		return
	}
	t.stop()
}

func (t *caller) StopChan() <-chan struct{} {
	return t.stopchan
}

func (t *caller) setError(e error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.firsterr == nil {
		t.firsterr = e
		t.stop()
	}
}

// Any calling methods must have the mutex acquired
func (t *caller) stop() {
	if !t.stopped {
		t.stopped = true
		close(t.stopchan)
	}
}

// Contra is a way to organize multiple structs with a `Start(...)` function
// which should be started and coordinated as a group
type Contra interface {
	Add(d ...Dancer)
	Start()
	Stop()
	Wait() error
}

type Dancer interface {
	Start(Follower) error
}

func NewContra() Contra {
	return &contra{
		members: make([]Dancer, 0),
		caller:  NewCaller(),
	}
}

type contra struct {
	members []Dancer
	caller  Caller
	mu      sync.Mutex
	started bool
}

func (c *contra) Add(d ...Dancer) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Do nothing if the contra is already started
	if c.started {
		return
	}
	c.members = append(c.members, d...)
}

func (c *contra) Start() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.started = true
	for _, d := range c.members {
		c.caller.Spawn(d.Start)
	}
}

func (c *contra) Stop() {
	c.caller.Stop()
}

func (c *contra) Wait() error {
	return c.caller.Wait()
}
