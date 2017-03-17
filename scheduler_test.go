package recurrent

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

type SchedulerSuite struct{}

func (s *SchedulerSuite) TestAutomaticPeriod(t *testing.T) {
	var (
		attempts  = 0
		clockChan = make(chan time.Time)
		clock     = newMockClock(clockChan, nil)
		sync      = make(chan struct{})
		done      = make(chan struct{})
	)

	defer close(sync)
	defer close(clockChan)

	sched := newSchedulerWithClock(
		time.Second,
		func() {
			attempts++
			sync <- struct{}{}
		},
		clock,
	)

	go func() {
		defer close(done)

		for i := 0; i < 25; i++ {
			clockChan <- time.Now()
			<-sync
		}
	}()

	<-done
	sched.Stop()
	Expect(attempts).To(Equal(25))
	Expect(clock.afterArgs[0]).To(Equal(time.Second))
}

func (s *SchedulerSuite) TestThrottledSchedule(t *testing.T) {
	var (
		attempts  = 0
		tickChan  = make(chan time.Time)
		ticker    = newMockTicker(tickChan)
		clockChan = make(chan time.Time)
		clock     = newMockClock(clockChan, ticker)
		sync      = make(chan struct{})
		done      = make(chan struct{})
	)

	defer close(sync)
	defer close(clockChan)

	sched := newThrottledSchedulerWithClock(
		time.Second,
		time.Millisecond,
		func() {
			attempts++
			sync <- struct{}{}
		},
		clock,
	)

	go func() {
		defer close(done)

		for i := 0; i < 25; i++ {
			clockChan <- time.Now()
			<-sync
		}
	}()

	// TODO - sync this thing better
	go func() {
		for i := 0; i < 40; i++ {
			select {
			case <-done:
				return
			case tickChan <- time.Now():
			}
		}
	}()

	<-done
	sched.Stop()
	Expect(attempts).To(Equal(25))
	Expect(clock.tickerArgs).To(HaveLen(1))
	Expect(clock.tickerArgs[0]).To(Equal(time.Millisecond))
}

func (s *SchedulerSuite) TestExplicitFire(t *testing.T) {
	var (
		attempts  = 0
		clockChan = make(chan time.Time)
		clock     = newMockClock(clockChan, nil)
		sync      = make(chan struct{})
		done      = make(chan struct{})
	)

	defer close(sync)
	defer close(clockChan)

	sched := newSchedulerWithClock(
		time.Second,
		func() {
			attempts++
			sync <- struct{}{}
		},
		clock,
	)

	go func() {
		defer close(done)

		for i := 0; i < 25; i++ {
			sched.Signal()
			<-sync
		}
	}()

	<-done
	sched.Stop()
	Expect(attempts).To(Equal(25))
}

func (s *SchedulerSuite) TestThrottledExplicitFire(t *testing.T) {
	var (
		attempts  = 0
		tickChan  = make(chan time.Time)
		ticker    = newMockTicker(tickChan)
		clockChan = make(chan time.Time)
		clock     = newMockClock(clockChan, ticker)
		sync      = make(chan struct{})
		done      = make(chan struct{})
	)

	defer close(sync)
	defer close(clockChan)

	sched := newThrottledSchedulerWithClock(
		time.Second,
		time.Millisecond,
		func() {
			attempts++
			sync <- struct{}{}
		},
		clock,
	)

	go func() {
		defer close(done)

		for i := 0; i < 100; i++ {
			sched.Signal()

			if i%4 == 0 {
				tickChan <- time.Now()
				<-sync
			}
		}

		for i := 0; i < 100; i++ {
			tickChan <- time.Now()
			clockChan <- time.Now()
			<-sync
		}
	}()

	<-done
	sched.Stop()
	Expect(attempts).To(Equal(125))
}

//
//
//

type mockClock struct {
	ch         <-chan time.Time
	ticker     ticker
	afterArgs  []time.Duration
	tickerArgs []time.Duration
}

type mockTicker struct {
	ch      chan time.Time
	stopped bool
}

func newMockClock(ch chan time.Time, ticker ticker) *mockClock {
	return &mockClock{
		ch:         ch,
		ticker:     ticker,
		afterArgs:  []time.Duration{},
		tickerArgs: []time.Duration{},
	}
}

func newMockTicker(ch chan time.Time) *mockTicker {
	return &mockTicker{
		ch:      ch,
		stopped: false,
	}
}

func (m *mockClock) After(duration time.Duration) <-chan time.Time {
	m.afterArgs = append(m.afterArgs, duration)
	return m.ch
}

func (m *mockClock) NewTicker(duration time.Duration) ticker {
	m.tickerArgs = append(m.tickerArgs, duration)
	return m.ticker
}

func (m *mockTicker) Chan() <-chan time.Time {
	return m.ch
}

func (m *mockTicker) Stop() {
	m.stopped = true
}
