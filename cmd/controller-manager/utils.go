// Copyright 2019 Chaos Mesh Authors.
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

package main

import (
	"time"
)

// Coalescer takes an input channel, and coalesced inputs with a timebound of interval.
// If input channel is closed, coalescer will signal one last time if we have any pending unsignalled events
// and close the output channel.
func Coalescer(interval time.Duration, input chan interface{}, stopCh <-chan struct{}) <-chan interface{} {
	output := make(chan interface{})
	go func() {
		var (
			signalled bool
			inputOpen = true // assume input chan is open before we run our select loop
		)
		setupLog.V(2).Info("Debouncing reconciliation signals with window",
			"interval", interval.String())
		for {
			select {
			case <-stopCh:
				return
			case <-time.After(interval):
				if signalled {
					setupLog.V(5).Info("Signalling reconciliation", "after interval", interval.String())
					output <- struct{}{}
					signalled = false
				}
			case _, inputOpen = <-input:
				if inputOpen { // only record events if the input channel is still open
					setupLog.V(4).Info("Got reconciliation signal", "interval", interval.String())
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
				setupLog.Info("Coalesce routine terminated, input channel is closed")
				close(output)

				return
			}
		}
	}()
	return output
}
