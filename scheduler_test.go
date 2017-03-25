package recurrent

import (
	"testing"
	"time"

	"github.com/efritz/glock"
	. "github.com/onsi/gomega"
)

type SchedulerSuite struct{}

func (s *SchedulerSuite) TestAutomaticPeriod(t *testing.T) {
	var (
		afterChan = make(chan time.Time)
		clock     = glock.NewMockClockWithAfterChan(afterChan)
		sync      = make(chan struct{})
		done      = make(chan struct{})
		attempts  = 0
	)

	defer close(sync)
	defer close(afterChan)

	scheduler := newSchedulerWithClock(
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
			afterChan <- time.Now()
			<-sync
		}
	}()

	scheduler.Start()
	<-done
	scheduler.Stop()
	Expect(attempts).To(Equal(25))
	Expect(clock.GetAfterArgs()[0]).To(Equal(time.Second))
}

func (s *SchedulerSuite) TestThrottledSchedule(t *testing.T) {
	var (
		afterChan  = make(chan time.Time)
		tickerChan = make(chan time.Time)
		clock      = glock.NewMockClockWithAfterChanAndTicker(afterChan, glock.NewMockTicker(tickerChan))
		sync       = make(chan struct{})
		done       = make(chan struct{})
		attempts   = 0
	)

	defer close(sync)
	defer close(afterChan)

	scheduler := newThrottledSchedulerWithClock(
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
			afterChan <- time.Now()
			<-sync
		}
	}()

	go func() {
		defer close(tickerChan)

		for {
			select {
			case <-done:
				return
			case tickerChan <- time.Now():
			}
		}
	}()

	scheduler.Start()
	<-done
	scheduler.Stop()
	Expect(attempts).To(Equal(25))

	tickerArgs := clock.GetTickerArgs()
	Expect(tickerArgs).To(HaveLen(1))
	Expect(tickerArgs[0]).To(Equal(time.Millisecond))
}

func (s *SchedulerSuite) TestExplicitFire(t *testing.T) {
	var (
		afterChan = make(chan time.Time)
		clock     = glock.NewMockClockWithAfterChan(afterChan)
		sync      = make(chan struct{})
		done      = make(chan struct{})
		attempts  = 0
	)

	defer close(sync)
	defer close(afterChan)

	scheduler := newSchedulerWithClock(
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
			scheduler.Signal()
			<-sync
		}
	}()

	scheduler.Start()
	<-done
	scheduler.Stop()
	Expect(attempts).To(Equal(25))
}

func (s *SchedulerSuite) TestThrottledExplicitFire(t *testing.T) {
	var (
		afterChan  = make(chan time.Time)
		tickerChan = make(chan time.Time)
		clock      = glock.NewMockClockWithAfterChanAndTicker(afterChan, glock.NewMockTicker(tickerChan))
		sync       = make(chan struct{})
		done       = make(chan struct{})
		attempts   = 0
	)

	defer close(sync)
	defer close(afterChan)
	defer close(tickerChan)

	scheduler := newThrottledSchedulerWithClock(
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
			scheduler.Signal()

			if i%4 == 0 {
				tickerChan <- time.Now()
				<-sync
			}
		}
	}()

	scheduler.Start()
	<-done
	scheduler.Stop()
	Expect(attempts).To(Equal(25))
}
