
# run

> A Go package to manage starting and graceful shutdown of servers and background tasks with timeout control.

---

## Overview

`run` provides a simple `Group` abstraction to coordinate the lifecycle of multiple components — starting them concurrently, waiting for any to fail or for cancellation, and gracefully stopping all with configurable timeouts.

It helps simplify common server orchestration scenarios like:
- Starting multiple services or goroutines
- Handling context cancellation or signals
- Enforcing startup and shutdown timeouts
- Collecting and combining errors during startup and shutdown

---

## Installation

```bash
go get github.com/yourusername/run
```

---

## Usage Example

```go
package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/run"
)

func main() {
	// Setup context that cancels on SIGINT or SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Create a new Group with optional start/stop timeout config
	g := run.NewGroup(
		run.WithStartTimeout(10 * time.Second),
		run.WithStopTimeout(5 * time.Second),
	)

	// Add a service start and stop functions
	g.Add(
		func() error {
			fmt.Println("Service started")
			// initialize service, return error on failure
			return nil
		},
		func(ctx context.Context) error {
			fmt.Println("Service stopping")
			// cleanup logic, honor context cancellation
			return nil
		},
	)

	// Run group, wait for start completion or failure, then wait for signal
	if err := g.Wait(ctx); err != nil {
		fmt.Printf("Run group error: %v\n", err)
	}
}
```

---

## API

- `NewGroup(opts ...Option) *Group`  
  Create a new run group with optional configurations.

- `(*Group) Add(start Start, stop Stop) *Group`  
  Add start and stop hooks. Start functions run concurrently; stop functions run in reverse order.

- `(*Group) Wait(ctx context.Context) error`  
  Start all hooks and wait for the first failure or external cancellation. Manages graceful shutdown.

- `WithStartTimeout(d time.Duration) Option`  
  Set the maximum allowed duration for all start functions.

- `WithStopTimeout(d time.Duration) Option`  
  Set the maximum allowed duration for all stop functions.

---

## Inspiration and References

- [oklog/run](https://github.com/oklog/run) — similar lifecycle management
- [golang.org/x/sync/errgroup](https://github.com/golang/sync/blob/master/errgroup/errgroup.go) — managing concurrent goroutines with error propagation
- [uber-go/fx](https://github.com/uber-go/fx) — dependency injection and lifecycle management framework
- [uber-go/guide](https://github.com/uber-go/guide) — best practices for Go

---
