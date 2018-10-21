# Squaredance

Simple task coordination

[![Build Status](https://travis-ci.org/lytics/squaredance.svg?branch=master)](https://travis-ci.org/lytics/squaredance) [![GoDoc](https://godoc.org/github.com/lytics/squaredance?status.svg)](https://godoc.org/github.com/lytics/squaredance)

Squaredance provides a clean interface for running groups of goroutines together, where any error should end the party.

Inspired by [Context](https://blog.golang.org/context), and [Tomb](https://gopkg.in/tomb.v1), squaredance embraces a single
way of handling these synchronization tasks. 

### Simple case
```go
c := squaredance.NewCaller()

vals := make(chan string)
// Start 5 tasks
for i := 0; i < 5; i++ {
	c.Spawn(func(sf squaredance.Follower) (err error) {
		for {
			select {
			case <-sf.StopChan():
				return
			case val := <-vals:
				fmt.Printf("%s\n", val)
			}
		}
	})

}

for _, v := range []string{"swing", "your", "partner", "round", "and", "round"} {
	vals <- v
}

c.Stop()
c.Wait()
```

All children of a `squaredance.Caller` are spawned as a single function call. If any of the functions returns an error,
the stop channel is closed, and the first error is returned from the `.Wait()` method.

Any of these functions can use the provided `squaredance.Follower` interfaces to listen to the stop channel, or to spawn more goroutines in the same group.

### Nested peer and return error
```go
c := squaredance.NewCaller()

c.Spawn(func(sf squaredance.Follower) error {
	val := make(chan string)
	sf.Spawn(func(sf squaredance.Follower) error {
		// This task returns immediately
		v := <-val
		fmt.Printf("Got a thing: %s\n", v)
		return fmt.Errorf("Fell down")
	})
	val <- "Spin"
	// This task waits until stop is called
	<-sf.StopChan()
	return nil
})

c.Stop()
err := c.Wait()
fmt.Printf("%v\n", err)
```
