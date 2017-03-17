package recurrent

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

type ChannelsSuite struct{}

func (s *ChannelsSuite) TestConvert(t *testing.T) {
	var (
		ch1 = make(chan time.Time)
		ch2 = convert(ch1)
	)

	go func() {
		for i := 1; i <= 200; i++ {
			ch1 <- time.Now().Add(time.Minute * time.Duration(-i))
		}
	}()

	for i := 1; i <= 200; i++ {
		Eventually(ch2, TIMEOUT, INTERVAL).Should(Receive(Equal(struct{}{})))
	}

	close(ch1)
	Eventually(ch2, TIMEOUT, INTERVAL).Should(BeClosed())
}

func (s *ChannelsSuite) TestThrottle(t *testing.T) {
	var (
		ch1 = make(chan struct{})
		ch2 = make(chan struct{})
		ch3 = throttle(ch1, ch2)
	)

	select {
	case ch2 <- struct{}{}:
		t.Errorf("did not expect an active reader")
	default:
	}

	ch1 <- struct{}{}

	go func() {
		ch2 <- struct{}{}
	}()

	Eventually(ch3, TIMEOUT, INTERVAL).Should(Receive(Equal(struct{}{})))
	close(ch1)
	Eventually(ch3, TIMEOUT, INTERVAL).Should(BeClosed())
}
