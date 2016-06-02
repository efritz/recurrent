package recurrent

import (
	"time"

	. "gopkg.in/check.v1"
)

const (
	duration = 64 * time.Millisecond
)

func (s *RecurrentSuite) TestAutomaticPeriod(c *C) {
	i := 0
	sched := NewScheduler(duration, func() {
		i++
	})

	<-time.After(duration*5 + duration/2)
	sched.Stop()
	c.Assert(i, Equals, 5)
}

func (s *RecurrentSuite) TestThrottledSchedule(c *C) {
	i := 0
	sched := NewThrottledScheduler(duration, duration/4, func() {
		i++
	})

	<-time.After(duration*5 + duration/2)
	sched.Stop()
	c.Assert(i, Equals, 5)
}

func (s *RecurrentSuite) TestExplicitFire(c *C) {
	i := 0
	sched := NewScheduler(duration, func() {
		i++
	})

	for j := 0; j < 25; j++ {
		sched.Signal()
		<-time.After(time.Millisecond)
	}

	<-time.After(duration*5 + duration/2)
	sched.Stop()
	c.Assert(i, Equals, 30)
}

func (s *RecurrentSuite) TestThrottledExplicitFire(c *C) {
	i := 0
	sched := NewThrottledScheduler(duration, duration/4, func() {
		i++
	})

	tick := time.NewTicker(duration / 8)
	defer tick.Stop()

	j := 0
	for _ = range tick.C {
		if j == 100 {
			break
		}

		sched.Signal()
		j++
	}

	sched.Stop()
	c.Assert(i, Equals, 50)
}

func (s *RecurrentSuite) TestStop(c *C) {
	i := 0
	sched := NewScheduler(duration, func() {
		i++
	})

	c.Assert(i, Equals, 0)
	sched.Signal()
	<-time.After(duration / 2)
	c.Assert(i, Equals, 1)
	sched.Stop()

	c.Assert(i, Equals, 1)
	sched.Signal()
	c.Assert(i, Equals, 1)
}

func (s *RecurrentSuite) TestSchedulerResetsAutomaticPeriod(c *C) {
	i := 0
	sched := NewScheduler(duration, func() {
		i++
	})

	c.Assert(i, Equals, 0)
	<-time.After(duration * 3 / 4)
	sched.Signal()
	<-time.After(duration * 1 / 4)
	c.Assert(i, Equals, 1)

	<-time.After(duration / 2)
	c.Assert(i, Equals, 1)
	<-time.After(duration / 2)
	c.Assert(i, Equals, 2)

	sched.Stop()
}
