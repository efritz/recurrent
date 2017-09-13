package recurrent

import (
	"time"

	"github.com/aphistic/sweet"
	"github.com/efritz/glock"
	. "github.com/onsi/gomega"
)

type ChanFactorySuite struct{}

func (s *ChanFactorySuite) TestHammerChanFactory(t sweet.T) {
	var (
		factory = newHammerChanFactory()
		ch      = factory.Chan()
	)

	for i := 1; i <= 200; i++ {
		<-ch
	}

	factory.Stop()
	eventually(ch).Should(BeClosed())
}

func (s *ChanFactorySuite) TestTickerChanFactory(t sweet.T) {
	var (
		clock   = glock.NewMockClock()
		ticker  = clock.NewTicker(time.Second)
		factory = newTickerChanFactory(ticker)
	)

	ch := factory.Chan()

	for i := 0; i < 100; i++ {
		clock.Advance(time.Second)
	}

	for i := 0; i < 100; i++ {
		Eventually(ch).Should(Receive())
	}

	Consistently(ch).ShouldNot(Receive())

	factory.Stop()
	eventually(ch).Should(BeClosed())
}
