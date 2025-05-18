package run_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/not-for-prod/run"
)

func ExampleGroup_Wait_basic() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	g := run.NewGroup()
	g.Add(func() error {
		return nil
	}, func(ctx context.Context) error {
		return nil
	})

	err := g.Wait(ctx)
	if err != nil {
		fmt.Println("error:", err)
	}
	// Output:
}

func ExampleGroup_Wait_startError() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	g := run.NewGroup()
	g.Add(func() error {
		return errors.New("start failed")
	}, func(ctx context.Context) error {
		return nil
	})

	err := g.Wait(ctx)
	if err != nil {
		fmt.Println(err)
	}
	// Output:
	// start failed
}

func ExampleGroup_Wait_startContextDeadlineExceeded() {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	g := run.NewGroup(run.WithStartTimeout(50 * time.Millisecond))
	g.Add(func() error {
		time.Sleep(100 * time.Millisecond)
		return nil
	}, func(ctx context.Context) error {
		return nil
	})

	err := g.Wait(ctx)
	if err != nil {
		fmt.Println(err)
	}
	// Output:
	// start context deadline exceeded
}

func ExampleGroup_Wait_stopError() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	g := run.NewGroup()
	g.Add(func() error {
		return errors.New("start failed")
	}, func(ctx context.Context) error {
		return errors.New("stop failed")
	})

	err := g.Wait(ctx)
	if err != nil {
		fmt.Println(err)
	}
	// Output:
	// start failed
	// stop failed
}

func ExampleGroup_Wait_stopContextDeadlineExceeded() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	g := run.NewGroup(run.WithStopTimeout(50 * time.Millisecond))
	g.Add(func() error {
		return errors.New("fail")
	}, func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	err := g.Wait(ctx)
	if err != nil {
		fmt.Println(err)
	}
	// Output:
	// fail
	// stop context deadline exceeded
}
