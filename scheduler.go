package recurrent

import "time"

type Scheduler struct {
	quit   chan bool
	signal chan bool
}

func NewScheduler(interval time.Duration, f func()) *Scheduler {
	sched := &Scheduler{
		quit:   make(chan bool),
		signal: make(chan bool, 1),
	}

	go func() {
		for {
			select {
			case <-time.After(interval):
				sched.Signal()

			case <-sched.signal:
				f()

			case <-sched.quit:
				break
			}
		}
	}()

	return sched
}

func NewThrottledScheduler(interval time.Duration, minInterval time.Duration, f func()) *Scheduler {
	sched := &Scheduler{
		quit:   make(chan bool),
		signal: make(chan bool, 1),
	}

	go func() {
		tick := time.NewTicker(minInterval)
		defer tick.Stop()

		t := throttle(tick.C, sched.signal)

		for {
			select {
			case <-time.After(interval):
				sched.Signal()

			case <-t:
				f()

			case <-sched.quit:
				break
			}
		}
	}()

	return sched
}

func (sched *Scheduler) Stop() {
	sched.quit <- true
}

func (sched *Scheduler) Signal() {
	select {
	case sched.signal <- true:

	default:
	}
}

func throttle(a <-chan time.Time, b chan bool) chan bool {
	c := make(chan bool)

	go func() {
		for _ = range a {
			c <- <-b
		}
	}()

	return c
}
