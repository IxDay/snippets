package runner

import (
	"context"
	"os"
	"os/signal"
	"sync"
)

type (
	Runner interface {
		Run(ctx context.Context) error
	}
	TeardownRunner interface {
		Run() error
		Teardown() error
	}
	RunnerFunc    func(context.Context) error
	RunnerManager []Runner
	Errors        []error
)

func (e Errors) Error() string { return e[0].Error() }

func (rf RunnerFunc) Run(ctx context.Context) error { return rf(ctx) }

func RunTeardownRunner(tr TeardownRunner) Runner {
	return TeardownRunnerFunc(tr.Run, tr.Teardown)
}

func Async(cb func() error) <-chan error {
	out := make(chan error)
	go func() { out <- cb() }()
	return out
}

func TeardownRunnerFunc(run func() error, teardown func() error) Runner {
	return RunnerFunc(func(ctx context.Context) error {
		select {
		case err := <-Async(run):
			return err
		case <-ctx.Done():
			return teardown()
		}
	})
}

func (rm RunnerManager) Wait(ctx context.Context) error {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(ctx)
	errors := []error{}
	defer cancel() // ensure that cancel has been called

	for _, runner := range rm {
		wg.Add(1)
		go func(runner Runner) {
			if err := runner.Run(ctx); err != nil {
				errors = append(errors, err)
			}
			wg.Done()
			cancel()
		}(runner)
	}

	wg.Wait()
	if len(errors) != 0 {
		return Errors(errors)
	} else {
		return nil
	}
}

func InterruptCb(cb func()) Runner {
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt)

	return RunnerFunc(func(ctx context.Context) error {
		select {
		case <-interrupt:
			cb()
			return nil
		case <-ctx.Done():
			return nil
		}
	})
}

func Interrupt() Runner { return InterruptCb(func() {}) }
