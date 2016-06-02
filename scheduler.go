package recurrent

import "time"

type Scheduler struct {
	target   func()
	interval time.Duration
	quit     chan bool
	signal   chan bool
}

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

func NewThrottledScheduler(interval time.Duration, minInterval time.Duration, target func()) *Scheduler {
	return newScheduler(interval, target, func(sched *Scheduler) {
		tick := time.NewTicker(minInterval)
		defer tick.Stop()

		sched.run(convert(tick.C))
	})
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
