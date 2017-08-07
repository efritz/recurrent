package recurrent

import (
	"time"

	"github.com/efritz/glock"
)

type (
	// Scheduler periodically executes the a target function. A scheduler can
	// be signaled to execute the function immediately as often as the user likes
	// via the Signal method (if the scheduler is configured to allow it).
	Scheduler interface {
		// Start the scheduler in a goroutine.
		Start()

		// Stop the scheduler. No additional signals are meaningful. This method
		// must not be called twice.
		Stop()

		// Signal the scheduler to execute the function immediately. This method is
		// always non-blocking, and may be ignored depending on if the scheduler is
		// throttling signals or not. This method must not be called after Stop.
		Signal()
	}

	scheduler struct {
		target   func()
		interval time.Duration
		clock    glock.Clock
		withChan func(f func(chan struct{}))
		quit     chan struct{}
		signal   chan struct{}
	}

	SchedulerConfig func(*scheduler)
)

func NewScheduler(target func(), configs ...SchedulerConfig) Scheduler {
	withChan := func(f func(chan struct{})) {
		quit := make(chan struct{})
		defer close(quit)

		f(hammer(quit))
	}

	scheduler := &scheduler{
		target:   target,
		interval: time.Second,
		clock:    glock.NewRealClock(),
		withChan: withChan,
		quit:     make(chan struct{}),
		signal:   make(chan struct{}, 1),
	}

	for _, config := range configs {
		config(scheduler)
	}

	return scheduler
}

func WithInterval(interval time.Duration) SchedulerConfig {
	return func(s *scheduler) { s.interval = interval }
}

func WithThrottle(minInterval time.Duration) SchedulerConfig {
	return func(s *scheduler) {
		s.withChan = func(f func(chan struct{})) {
			ticker := s.clock.NewTicker(minInterval)
			defer ticker.Stop()

			f(convert(ticker.Chan()))
		}
	}
}

func withClock(clock glock.Clock) SchedulerConfig {
	return func(s *scheduler) { s.clock = clock }
}

func (s *scheduler) Start() {
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

func (s *scheduler) Stop() {
	close(s.quit)
}

func (s *scheduler) Signal() {
	select {
	case s.signal <- struct{}{}:
	default:
	}
}
