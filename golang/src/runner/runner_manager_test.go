package runner

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

var (
	interrupted   = fmt.Errorf("interrupted")
	uninterrupted = fmt.Errorf("uninterrupted")
	misc          = fmt.Errorf("misc")
)

func CheckErrors(t *testing.T, got, expected error) {
	t.Helper()
	ok := (got == nil && expected == nil) ||
		(got != nil && expected != nil && got.Error() == expected.Error())

	if !ok {
		t.Errorf("unexpected error got: %q, expected: %q", got, expected)
	}
}

// stub runner which return interrupt error if interrupted, and uninterrupted if gone through
func stubRunner() (start, stop chan bool, runner Runner) {
	start, stop = make(chan bool), make(chan bool)

	runner = RunnerFunc(func(ctx context.Context) error {
		close(start)
		select {
		case <-ctx.Done():
			return interrupted
		case <-Async(func() error {
			time.Sleep(5 * time.Millisecond)
			return nil
		}):
			return uninterrupted
		}
	})
	return
}

// works most of the time, sometimes scheduling is working a bit differently,
// did not manage to get this right, but anyway 90% success proves the point
func TestInterrupt(t *testing.T) {
	var err error
	var entries = []struct {
		cb  func()
		err error
	}{
		{func() {}, uninterrupted},
		{
			func() {
				p, err := os.FindProcess(os.Getpid())
				CheckErrors(t, err, nil)
				CheckErrors(t, p.Signal(os.Interrupt), nil)
			},
			interrupted,
		},
	}
	for _, entry := range entries {
		start, stop, runner := stubRunner()
		rm := RunnerManager{Interrupt(), runner}

		go func() {
			err = rm.Wait(context.Background())
			close(stop)
		}()
		<-start
		entry.cb()
		<-stop
		CheckErrors(t, err, entry.err)
	}
}

func TestRunnerManager(t *testing.T) {
	entries := []struct {
		out error
		err error
	}{
		{nil, nil},
		{misc, misc},
	}
	for _, entry := range entries {
		rm := RunnerManager{RunnerFunc(func(_ context.Context) error { return entry.out })}
		want, got := entry.err, rm.Wait(context.Background())
		CheckErrors(t, got, want)
	}
}

type stubTeardownRunner struct {
	run      bool
	teardown bool
}

func (str *stubTeardownRunner) Run() error {
	time.Sleep(5 * time.Millisecond)
	str.run = true
	return nil
}
func (str *stubTeardownRunner) Teardown() error {
	str.teardown = true
	return nil
}

func TestRunTeardownRunner(t *testing.T) {
	var stub *stubTeardownRunner
	var errRunner = RunnerFunc(func(_ context.Context) error { return nil })
	var errString = "TeardownRunner.Run(%q) = %t (run), %t (teardown) want: %t, %t"

	// test without stopping (teardown not called)
	stub = &stubTeardownRunner{}
	CheckErrors(t, RunnerManager{RunTeardownRunner(stub)}.Wait(context.Background()), nil)
	if !stub.run || stub.teardown {
		t.Errorf(errString, "no stop", stub.run, stub.teardown, true, false)
	}

	// test with stopping (teardown called)
	stub = &stubTeardownRunner{}
	CheckErrors(t, RunnerManager{errRunner, RunTeardownRunner(stub)}.Wait(context.Background()), nil)
	if stub.run || !stub.teardown {
		t.Errorf(errString, "stop", stub.run, stub.teardown, false, true)
	}
}
