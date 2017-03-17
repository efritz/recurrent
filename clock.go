package recurrent

import "time"

type (
	clock interface {
		After(duration time.Duration) <-chan time.Time
		NewTicker(duration time.Duration) ticker
	}

	ticker interface {
		Chan() <-chan time.Time
		Stop()
	}

	realClock  struct{}
	realTicker struct{ *time.Ticker }
)

func (rc *realClock) After(duration time.Duration) <-chan time.Time {
	return time.After(duration)
}

func (rc *realClock) NewTicker(duration time.Duration) ticker {
	return &realTicker{time.NewTicker(duration)}
}

func (rt *realTicker) Chan() <-chan time.Time {
	return rt.C
}

func (rt *realTicker) Stop() {
	rt.Stop()
}
