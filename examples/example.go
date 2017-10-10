package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/efritz/recurrent"
)

var target = func() {
	fmt.Printf("Hello!\n")
}

func makeSignalChan() chan struct{} {
	ch := make(chan struct{})

	go func() {
		fmt.Printf("Press enter to signal\n")
		reader := bufio.NewReader(os.Stdin)

		for {
			if _, err := reader.ReadString('\n'); err != nil {
				panic(err.Error())
			}

			ch <- struct{}{}
		}
	}()

	return ch
}

func makeThrottledScheduler() recurrent.Scheduler {
	return recurrent.NewScheduler(
		target,
		recurrent.WithInterval(time.Second*5),
		recurrent.WithThrottle(time.Second/2),
	)
}

func makeUnthrottledScheduler() recurrent.Scheduler {
	return recurrent.NewScheduler(
		target,
		recurrent.WithInterval(time.Second),
	)
}

func makeScheduler() recurrent.Scheduler {
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
