package recurrent

import (
	"testing"
	"time"

	"github.com/aphistic/sweet"
	"github.com/aphistic/sweet-junit"
	"github.com/efritz/glock"
	. "github.com/onsi/gomega"
)

const (
	TIMEOUT  = time.Millisecond * 1
	INTERVAL = time.Millisecond / 100
)

func TestMain(m *testing.M) {
	RegisterFailHandler(sweet.GomegaFail)

	sweet.Run(m, func(s *sweet.S) {
		s.RegisterPlugin(junit.NewPlugin())

		s.AddSuite(&SchedulerSuite{})
		s.AddSuite(&ChanFactorySuite{})
	})
}

type SchedulerSuite struct{}

func (s *SchedulerSuite) TestAutomaticPeriod(t sweet.T) {
	var (
		clock    = glock.NewMockClock()
		sync     = make(chan struct{})
		done     = make(chan struct{})
		attempts = 0
	)

	defer close(sync)

	scheduler := NewScheduler(
		func() {
			attempts++
			sync <- struct{}{}
		},
		WithInterval(time.Second),
		withClock(clock),
	)

	go func() {
		defer close(done)

		for i := 0; i < 25; i++ {
			clock.BlockingAdvance(time.Second)
			<-sync
		}
	}()

	scheduler.Start()
	<-done
	scheduler.Stop()
	Expect(attempts).To(Equal(25))
	Expect(clock.GetAfterArgs()[0]).To(Equal(time.Second))
}

func (s *SchedulerSuite) TestThrottledSchedule(t sweet.T) {
	var (
		clock    = glock.NewMockClock()
		sync     = make(chan struct{})
		done     = make(chan struct{})
		attempts = 0
	)

	scheduler := NewScheduler(
		func() {
			attempts++
			sync <- struct{}{}
		},
		WithInterval(time.Second),
		WithThrottle(time.Millisecond),
		withClock(clock),
	)

	go func() {
		defer close(done)

		for i := 0; i < 25; i++ {
			clock.BlockingAdvance(time.Second)
			<-sync
		}
	}()

	scheduler.Start()
	<-done
	scheduler.Stop()
	Expect(attempts).To(Equal(25))
}

func (s *SchedulerSuite) TestExplicitFire(t sweet.T) {
	var (
		clock    = glock.NewMockClock()
		sync     = make(chan struct{})
		done     = make(chan struct{})
		attempts = 0
	)

	defer close(sync)

	scheduler := NewScheduler(
		func() {
			attempts++
			sync <- struct{}{}
		},
		WithInterval(time.Second),
		withClock(clock),
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

func (s *SchedulerSuite) TestThrottledExplicitFire(t sweet.T) {
	var (
		clock    = glock.NewMockClock()
		sync     = make(chan struct{})
		done     = make(chan struct{})
		attempts = 0
	)

	defer close(sync)

	scheduler := NewScheduler(
		func() {
			attempts++
			sync <- struct{}{}
		},
		WithInterval(time.Second),
		WithThrottle(time.Millisecond),
		withClock(clock),
	)

	go func() {
		defer close(done)

		for i := 0; i < 100; i++ {
			scheduler.Signal()

			if i%4 == 0 {
				clock.BlockingAdvance(time.Second)
				<-sync
			}
		}
	}()

	scheduler.Start()
	<-done
	scheduler.Stop()
	Expect(attempts).To(Equal(25))
}

func (s *SchedulerSuite) TestThrottle(t sweet.T) {
	var (
		ch1 = make(chan struct{})
		ch2 = make(chan struct{})
		ch3 = throttle(ch1, ch2)
	)

	Consistently(ch2).ShouldNot(Receive())
	ch1 <- struct{}{}

	go func() {
		ch2 <- struct{}{}
	}()

	Eventually(ch3).Should(Receive(Equal(struct{}{})))
	close(ch1)
	Eventually(ch3).Should(BeClosed())
}
