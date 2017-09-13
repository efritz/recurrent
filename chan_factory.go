package recurrent

import "github.com/efritz/glock"

type (
	chanFactory interface {
		Chan() <-chan struct{}
		Stop()
	}

	hammerChanFactory struct {
		quit chan struct{}
	}

	tickerChanFactory struct {
		ticker glock.Ticker
		quit   chan struct{}
	}
)

func newHammerChanFactory() chanFactory {
	return &hammerChanFactory{
		quit: make(chan struct{}),
	}
}

func (cf *hammerChanFactory) Chan() <-chan struct{} {
	ch := make(chan struct{})

	go func() {
		defer close(ch)

		for {
			select {
			case ch <- struct{}{}:
			case <-cf.quit:
				return
			}
		}
	}()

	return ch
}

func (cf *hammerChanFactory) Stop() {
	close(cf.quit)
}

func newTickerChanFactory(ticker glock.Ticker) chanFactory {
	return &tickerChanFactory{
		ticker: ticker,
		quit:   make(chan struct{}),
	}
}

func (cf *tickerChanFactory) Chan() <-chan struct{} {
	ch := make(chan struct{})

	go func() {
		defer close(ch)

		tickerChan := cf.ticker.Chan()
		defer cf.ticker.Stop()

		for {
			select {
			case <-tickerChan:
			case <-cf.quit:
				return
			}

			select {
			case ch <- struct{}{}:
			case <-cf.quit:
				return
			}
		}
	}()

	return ch
}

func (cf *tickerChanFactory) Stop() {
	close(cf.quit)
}
