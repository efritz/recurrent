package recurrent

import "time"

// Scheduler is a struct that holds the target function and its schedule
// configuration.
type Scheduler struct {
	// User parameters
	target   func()
	interval time.Duration
	clock    clock

	// Control channels
	quit   chan struct{}
	signal chan struct{}
}

// NewScheduler creates a new scheduler which periodically executes the given
// function. This scheduler can be signaled to execute the function immediately
// as often as the user likes.
func NewScheduler(interval time.Duration, target func()) *Scheduler {
	return newSchedulerWithClock(interval, target, &realClock{})
}

// NewThrottledScheduler creates a new scheduler which periodically executes
// the given function. The minInterval parameter specifies the frequency at
// which the signals for immediate execution are allowed. Any additional signals
// within the interval are ignored.
func NewThrottledScheduler(interval time.Duration, minInterval time.Duration, target func()) *Scheduler {
	return newThrottledSchedulerWithClock(interval, minInterval, target, &realClock{})
}

func newSchedulerWithClock(interval time.Duration, target func(), clock clock) *Scheduler {
	return newScheduler(interval, target, clock, func(sched *Scheduler) {
		c := make(chan struct{})
		q := make(chan struct{})
		defer close(q)

		go func() {
			defer close(c)

			for {
				select {
				case c <- struct{}{}:
				case <-q:
					return
				}
			}
		}()

		// TODO - instead, just return the channel
		sched.run(c)
	})
}

func newThrottledSchedulerWithClock(interval time.Duration, minInterval time.Duration, target func(), clock clock) *Scheduler {
	return newScheduler(interval, target, clock, func(sched *Scheduler) {
		ticker := clock.NewTicker(minInterval)
		defer ticker.Stop()

		// TODO - instead, just return the channel
		sched.run(convert(ticker.Chan()))
	})
}

func newScheduler(interval time.Duration, target func(), clock clock, init func(*Scheduler)) *Scheduler {
	s := &Scheduler{
		target:   target,
		interval: interval,
		clock:    clock,
		quit:     make(chan struct{}),
		signal:   make(chan struct{}, 1), // TODO - see if we can not do this
	}

	go init(s)
	return s
}

// Run the scheduler. While the scheduler is running, receive a value from either
// the inteval ticker or the channel "t", which throttles the signal channel with
// respect to the channel "c". If we read from the interval channel, we write to
// the signal channel (or no-op if the channel already has a value). Otherwise
// we read from the "t" channel and want to acutally execute the target function.
// In either case, we "reset" the interval ticker by making a new fires-once time
// channel.
func (s *Scheduler) run(c chan struct{}) {
	defer close(s.signal)

	t := throttle(c, s.signal)

	for {
		select {
		case <-t:
			s.target()

		case <-s.clock.After(s.interval):
			s.Signal()

		case <-s.quit:
			return
		}
	}
}

}

// Stop the scheduler. No additional signals are meaningful. This method
// must not be called twice.
func (s *Scheduler) Stop() {
	close(s.quit)
}

// Signal the scheduler to execute the function immediately. This method is
// always non-blocking, and may be ignored depending on if the scheduler is
// throttling signals or not. This method must not be called after Stop.
func (s *Scheduler) Signal() {
	select {
	case s.signal <- struct{}{}:
	default:
	}
}
