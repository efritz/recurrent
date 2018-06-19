package recurrent

import (
	"sync"
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

		// Stop the scheduler.
		Stop()

		// Signal the scheduler to execute the function immediately. This method is
		// always non-blocking, and may be ignored depending on if the scheduler is
		// throttling signals or not. This method must only be called while the
		// scheduler is active.
		Signal()

		// Reset the interval timer in the scheduler. This method must only be called
		// while the scheduler is active.
		Reset()
	}

	scheduler struct {
		target    func()
		interval  time.Duration
		clock     glock.Clock
		factory   chanFactory
		skipFirst bool
		quit      chan struct{}
		signal    chan struct{}
		reset     chan struct{}
		wg        sync.WaitGroup
		once      *sync.Once
	}

	// ConfigFunc is a function used to initialize a new scheduler.
	ConfigFunc func(*scheduler)
)

// NewScheduler creates a new scheduler that will invoke the target function.
func NewScheduler(target func(), configs ...ConfigFunc) Scheduler {
	scheduler := &scheduler{
		target:    target,
		interval:  time.Second,
		skipFirst: false,
		clock:     glock.NewRealClock(),
		factory:   newHammerChanFactory(),
		quit:      make(chan struct{}),
		signal:    make(chan struct{}, 1),
		reset:     make(chan struct{}),
		once:      &sync.Once{},
	}

	for _, config := range configs {
		config(scheduler)
	}

	return scheduler
}

// WithInterval sets the interval at which the scheduler will invoke the
// scheduled function (default is one second).
func WithInterval(interval time.Duration) ConfigFunc {
	return func(s *scheduler) {
		s.interval = interval
	}
}

// WithThrottle sets the minimum duration between two invocations of the
// scheduled function (there is no default minimum).
func WithThrottle(minInterval time.Duration) ConfigFunc {
	return func(s *scheduler) {
		s.factory = newTickerChanFactory(s.clock.NewTicker(minInterval))
	}
}

// WithClock sets the clock used by the scheduler (useful for testing).
func WithClock(clock glock.Clock) ConfigFunc {
	return func(s *scheduler) {
		s.clock = clock
	}
}

// WithSkipFirstInvocation will flag the first invocation (immediately after
// Start is called) to be skipped. The first invocation will be after the
// first call to Signal, Reset, or after the initial timeout period elapses.
func WithSkipFirstInvocation() ConfigFunc {
	return func(s *scheduler) {
		s.skipFirst = true
	}
}

func (s *scheduler) Start() {
	s.wg.Add(2)

	go func() {
		defer s.wg.Done()

		for {
			select {
			case <-s.clock.After(s.interval):
				s.Signal()

			case <-s.reset:
				continue

			case <-s.quit:
				return
			}
		}
	}()

	go func() {
		defer s.wg.Done()

		ch := s.factory.Chan()
		defer s.factory.Stop()

		tick := throttle(ch, s.signal)

		for {
			select {
			case <-tick:
				s.target()

			case <-s.quit:
				return
			}
		}
	}()

	if !s.skipFirst {
		s.Signal()
	}
}

func (s *scheduler) Stop() {
	s.once.Do(func() {
		close(s.quit)
		s.wg.Wait()
		close(s.signal)
		close(s.reset)
	})
}

func (s *scheduler) Signal() {
	select {
	case s.signal <- struct{}{}:
	default:
	}
}

func (s *scheduler) Reset() {
	s.reset <- struct{}{}
}

func throttle(ch1 <-chan struct{}, ch2 <-chan struct{}) <-chan struct{} {
	ch3 := make(chan struct{})

	go func() {
		defer close(ch3)

		for range ch1 {
			ch3 <- <-ch2
		}
	}()

	return ch3
}
