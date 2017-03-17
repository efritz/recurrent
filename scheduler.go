package recurrent

import "time"

type (
	// Scheduler is a struct that holds the target function and its schedule
	// configuration.
	Scheduler struct {
		target   func()
		interval time.Duration
		clock    clock
		withChan func(f func(chan struct{}))

		// Control channels
		quit   chan struct{}
		signal chan struct{}
	}
)

// NewScheduler creates a new scheduler which periodically executes the given
// function. This scheduler can be signaled to execute the function immediately
// as often as the user likes via the Signal method.
func NewScheduler(interval time.Duration, target func()) *Scheduler {
	return newSchedulerWithClock(interval, target, &realClock{})
}

// NewThrottledScheduler creates a new scheduler which periodically executes
// the given function. The minInterval parameter specifies the frequency at
// which the signals for immediate execution are allowed. Additional signals
// within the interval are ignored.
func NewThrottledScheduler(interval time.Duration, minInterval time.Duration, target func()) *Scheduler {
	return newThrottledSchedulerWithClock(interval, minInterval, target, &realClock{})
}

func newSchedulerWithClock(interval time.Duration, target func(), clock clock) *Scheduler {
	return newScheduler(interval, target, clock, func(f func(chan struct{})) {
		quit := make(chan struct{})
		defer close(quit)

		f(hammer(quit))
	})
}

func newThrottledSchedulerWithClock(interval time.Duration, minInterval time.Duration, target func(), clock clock) *Scheduler {
	return newScheduler(interval, target, clock, func(f func(chan struct{})) {
		ticker := clock.NewTicker(minInterval)
		defer ticker.Stop()

		f(convert(ticker.Chan()))
	})
}

func newScheduler(interval time.Duration, target func(), clock clock, withChan func(f func(chan struct{}))) *Scheduler {
	return &Scheduler{
		target:   target,
		interval: interval,
		clock:    clock,
		withChan: withChan,
		quit:     make(chan struct{}),
		signal:   make(chan struct{}, 1),
	}
}

// Start the scheduler in a goroutine.
func (s *Scheduler) Start() {
	go func() {
		defer close(s.signal)

		s.withChan(func(c chan struct{}) {
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
		})
	}()
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
