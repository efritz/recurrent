package recurrent

import "time"

type Scheduler struct {
	// User parameters
	target   func()
	interval time.Duration

	// Control channels
	quit   chan bool
	signal chan bool
}

// Create a new scheduler which periodically executes the given function.
// This scheduler can be signaled to execute the function immediately as
// often as the user likes.
func NewScheduler(interval time.Duration, target func()) *Scheduler {
	return newScheduler(interval, target, func(sched *Scheduler) {
		c := make(chan bool)
		q := make(chan bool)

		go func() {
			for {
				select {
				case <-q:
					break

				case c <- true:
				}
			}

			close(c)
		}()

		sched.run(c)
		q <- true
	})
}

// Create a new scheduler which periodically executes the given function.
// The minInterval parameter specifies the frequency at which the signals
// for immediate execution are allowed. Any additional signals within the
// interval are ignored.
func NewThrottledScheduler(interval time.Duration, minInterval time.Duration, target func()) *Scheduler {
	return newScheduler(interval, target, func(sched *Scheduler) {
		tick := time.NewTicker(minInterval)
		defer tick.Stop()

		sched.run(convert(tick.C))
	})
}

// Stop the scheduler. No additional signals are meaningful.
func (sched *Scheduler) Stop() {
	sched.quit <- true
}

// Signal the scheduler to execute the function immediately. This method is
// always non-blocking, and may be ignored depending on if the scheduler is
// throttling signals or not.
func (sched *Scheduler) Signal() {
	select {
	case sched.signal <- true:

	default:
	}
}

func newScheduler(interval time.Duration, target func(), init func(*Scheduler)) *Scheduler {
	sched := &Scheduler{
		target:   target,
		interval: interval,
		quit:     make(chan bool),
		signal:   make(chan bool, 1),
	}

	go init(sched)
	return sched
}

// Run the scheduler. While the scheduler is running, receive a value from either
// the inteval ticker or the channel "t", which throttles the signal channel with
// respect to the channel "c". If we read from the interval channel, we send true
// on the signal channel (or no-op if the channel already has a value). Otherwise
// we read from the "t" channel and want to acutally execute the target function.
// In either case, we "reset" the interval ticker by making a new fires-once time
// channel.
func (sched *Scheduler) run(c chan bool) {
	t := throttle(c, sched.signal)

	for {
		select {
		case <-time.After(sched.interval):
			sched.Signal()

		case <-t:
			sched.target()

		case <-sched.quit:
			break
		}
	}
}

// Convert a ticker channel to a bool channel.
func convert(a <-chan time.Time) chan bool {
	c := make(chan bool)

	go func() {
		for _ = range a {
			c <- true
		}

		close(c)
	}()

	return c
}

// Create a channel which pipes contents from channel "b" but only
// at the rate of channel "a".
func throttle(a chan bool, b chan bool) chan bool {
	c := make(chan bool)

	go func() {
		for _ = range a {
			c <- <-b
		}

		close(c)
	}()

	return c
}
