// Copyright 2019 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("util")

// Coalescer takes an input chan, and coalesced inputs with a timebound of interval, after which
// it signals on output chan with the last value from input chan
func Coalescer(interval time.Duration, input chan interface{}, stopCh <-chan struct{}) <-chan interface{} {
	output := make(chan interface{})
	go func() {
		var (
			signalled bool
			inputOpen = true // assume input chan is open before we run our select loop
		)
		log.V(2).Info("debouncing reconciliation signals with window",
			"interval", interval.String())
		for {
			select {
			case <-stopCh:
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
