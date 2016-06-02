package recurrent

import "time"

type Scheduler struct {
	target   func()
	interval time.Duration
	quit     chan bool
	signal   chan bool
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
