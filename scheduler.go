package recurrent

import "time"

type Scheduler struct {
	// User parameters
	target   func()
	interval time.Duration

	// Control channels
	quit   chan struct{}
	signal chan struct{}
}

// Create a new scheduler which periodically executes the given function.
// This scheduler can be signaled to execute the function immediately as
// often as the user likes.
func NewScheduler(interval time.Duration, target func()) *Scheduler {
	return newScheduler(interval, target, func(sched *Scheduler) {
		c := make(chan struct{})
		q := make(chan struct{})

		go func() {
			for {
				select {
				case <-q:
					break

				case c <- struct{}{}:
				}
			}

			close(c)
			close(q)
		}()

		sched.run(c)
		q <- struct{}{}
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
	sched.quit <- struct{}{}
}

// Signal the scheduler to execute the function immediately. This method is
// always non-blocking, and may be ignored depending on if the scheduler is
// throttling signals or not.
func (sched *Scheduler) Signal() {
	select {
	case sched.signal <- struct{}{}:

	default:
	}
}

func newScheduler(interval time.Duration, target func(), init func(*Scheduler)) *Scheduler {
	sched := &Scheduler{
		target:   target,
		interval: interval,
		quit:     make(chan struct{}),
		signal:   make(chan struct{}, 1),
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
func (sched *Scheduler) run(c chan struct{}) {
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

	close(sched.signal)
}

// Convert a ticker channel to a struct{} channel.
func convert(a <-chan time.Time) chan struct{} {
	c := make(chan struct{})

	go func() {
		for _ = range a {
			c <- struct{}{}
		}

		close(c)
	}()

	return c
}

// Create a channel which pipes contents from channel "b" but only
// at the rate of channel "a".
func throttle(a chan struct{}, b chan struct{}) chan struct{} {
	c := make(chan struct{})

	go func() {
		for _ = range a {
			c <- <-b
		}

		close(c)
	}()

	return c
}
