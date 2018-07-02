package main

import (
	"context"
	"fmt"
	"time"

	"runner"
)

func Misc() runner.Runner {
	return runner.RunnerFunc(func(ctx context.Context) error {
		fmt.Println("Starting doing something...")
		select {
		case <-time.After(5 * time.Second):
			fmt.Println("Done doing something...")
			return nil
		case <-ctx.Done():
			fmt.Println("Tearing down...")
			time.Sleep(2 * time.Second)
			fmt.Println("Tearing down, finished...")
			return nil
		}
	})
}

func main() {
	runner.RunnerManager{
		runner.InterruptCb(func() { fmt.Printf("\nCaught interrupt, aborting...\n") }),
		Misc(),
	}.Wait(context.Background())
}
