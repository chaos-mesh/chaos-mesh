package utils

import (
	"context"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("util")

// Coalesce takes an input chan, and coalesced inputs with a timebound of interval, after which
// it signals on output chan with the last value from input chan
func Coalesce(ctx context.Context, interval time.Duration, input chan interface{}) <-chan interface{} {
	output := make(chan interface{})
	go func() {
		var (
			signalled bool
			inputOpen = true // assume input chan is open before we run our select loop
		)
		log.V(2).Info("debouncing reconciliation signals with window",
			"interval", interval.String())
		for {
			doneCh := ctx.Done()
			select {
			case <-doneCh:
				if signalled {
					output <- struct{}{}
				}
				return
			case <-time.After(interval):
				if signalled {
					log.V(5).Info("signalling reconciliation", "after interval", interval.String())
					output <- struct{}{}
					signalled = false
				}
			case _, inputOpen = <-input:
				if inputOpen { // only record events if the input channel is still open
					log.V(4).Info("got reconciliation signal", "interval", interval.String())
					signalled = true
				}
			}
			// stop running the Coalescer only when all input+output channels are closed!
			if !inputOpen {
				// input is closed, so lets signal one last time if we have any pending unsignalled events
				if signalled {
					// send final event, so we dont miss the trailing event after input chan close
					output <- struct{}{}
				}
				log.Info("coalesce routine terminated, input channel is closed")
				return
			}
		}
	}()
	return output
}
