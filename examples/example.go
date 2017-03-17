package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/efritz/recurrent"
)

var (
	target = func() {
		fmt.Printf("Hello!\n")
	}
)

func makeSignalChan() chan struct{} {
	ch := make(chan struct{})

	go func() {
		fmt.Printf("Press enter to signal\n")
		reader := bufio.NewReader(os.Stdin)

		for {
			reader.ReadString('\n')
			ch <- struct{}{}
		}
	}()

	return ch
}

func makeThrottledScheduler() *recurrent.Scheduler {
	return recurrent.NewThrottledScheduler(time.Second*5, time.Second/2, target)
}

func makeUnthrottledScheduler() *recurrent.Scheduler {
	return recurrent.NewScheduler(time.Second, target)
}

func makeScheduler() *recurrent.Scheduler {
	if os.Getenv("THROTTLED") != "" {
		return makeThrottledScheduler()
	}

	return makeUnthrottledScheduler()
}

func main() {
	var (
		signals   = makeSignalChan()
		scheduler = makeScheduler()
	)

	scheduler.Start()
	defer scheduler.Stop()

	for range signals {
		scheduler.Signal()
	}
}
