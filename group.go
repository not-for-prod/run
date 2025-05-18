package run

import (
	"context"
	"errors"
	"sync"
)

var (
	// ErrStartContextDeadlineExceeded is returned when the start phase exceeds the configured timeout.
	ErrStartContextDeadlineExceeded = errors.New("start context deadline exceeded")

	// ErrStopContextDeadlineExceeded is returned when the stop phase exceeds the configured timeout.
	ErrStopContextDeadlineExceeded = errors.New("stop context deadline exceeded")
)

// Start is a function that initializes a component. It should return quickly or return an error.
type Start func() error

// Stop is a function that gracefully shuts down a component using the provided context.
type Stop func(ctx context.Context) error

// Group manages the coordinated startup and shutdown of multiple components.
type Group struct {
	opts     options // configuration options (e.g., timeouts)
	mu       sync.Mutex
	starters []Start // registered start functions
	stoppers []Stop  // registered stop functions
}

// NewGroup creates a new Group with the given options.
func NewGroup(options ...Option) *Group {
	opts := defaultOptions
	for _, opt := range options {
		opt.apply(&opts)
	}
	return &Group{opts: opts}
}

// Add registers a start and stop function to the group.
//
// Start is called during Group.Wait to initialize the component.
// Stop is called during shutdown or if any Start function fails.
func (g *Group) Add(start Start, stop Stop) *Group {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.starters = append(g.starters, start)
	g.stoppers = append(g.stoppers, stop)
	return g
}

// Wait starts all registered components, waits for completion or error,
// and ensures stop functions are called in reverse order.
//
// The behavior is as follows:
// 1. Starts all components concurrently within a start timeout.
// 2. If any start fails, calls all stop functions.
// 3. If start times out, calls stop functions and returns a timeout error.
// 4. If all components start successfully, blocks until ctx is canceled, then stops.
func (g *Group) Wait(ctx context.Context) error {
	startCtx, startCancel := context.WithTimeout(ctx, g.opts.startTimeout)
	defer startCancel()

	var wg sync.WaitGroup
	startErrors := make(chan error, len(g.starters))

	// Start all registered Start functions concurrently.
	for _, start := range g.starters {
		wg.Add(1)
		go func(a Start) {
			defer wg.Done()
			if err := a(); err != nil {
				startErrors <- err
			}
		}(start)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(startErrors)
		close(done)
	}()

	select {
	case <-ctx.Done():
		// External context canceled — stop components.
		return g.stop()

	case <-startCtx.Done():
		// Start phase timed out — stop components and return timeout error.
		err := g.stop()
		if err != nil {
			return errors.Join(ErrStartContextDeadlineExceeded, err)
		}
		return ErrStartContextDeadlineExceeded

	case <-done:
		// All starters completed, now check for any errors.
		var errs []error
		for err := range startErrors {
			errs = append(errs, err)
		}
		if len(errs) > 0 {
			stopErr := g.stop()
			if stopErr != nil {
				errs = append(errs, stopErr)
			}
			return errors.Join(errs...)
		}

		// Successful start — wait for external signal to stop.
		<-ctx.Done()
		return g.stop()
	}
}

// stop shuts down all registered components in reverse order.
//
// Stops run concurrently within a stop timeout.
// Errors from any stop function are collected and returned.
func (g *Group) stop() error {
	stopCtx, stopCancel := context.WithTimeout(context.Background(), g.opts.stopTimeout)
	defer stopCancel()

	var wg sync.WaitGroup
	stopErrors := make(chan error, len(g.stoppers))

	// Stop in reverse order of Add
	for i := len(g.stoppers) - 1; i >= 0; i-- {
		stopper := g.stoppers[i]
		wg.Add(1)
		go func(a Stop) {
			defer wg.Done()
			if err := a(stopCtx); err != nil {
				stopErrors <- err
			}
		}(stopper)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
		close(stopErrors)
	}()

	var errs []error

	// Wait for stop to complete or timeout
	select {
	case <-stopCtx.Done():
		errs = append(errs, ErrStopContextDeadlineExceeded)
	case <-done:
	}

	// Collect stop errors
	for err := range stopErrors {
		errs = append(errs, err)
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}
