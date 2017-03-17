package recurrent

import "time"

// Convert a ticker channel to a struct{} channel.
func convert(ch1 <-chan time.Time) chan struct{} {
	ch2 := make(chan struct{})

	go func() {
		defer close(ch2)

		for _ = range ch1 {
			ch2 <- struct{}{}
		}
	}()

	return ch2
}

// Create a channel which receives a value once a value is ready
// from both ch1 and ch2. The channel is closed once ch1 closes.
func throttle(ch1 chan struct{}, ch2 chan struct{}) chan struct{} {
	ch3 := make(chan struct{})

	go func() {
		defer close(ch3)

		for _ = range ch1 {
			ch3 <- <-ch2
		}
	}()

	return ch3
}
