/*
Copyright 2026 The OpenYurt Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package engine

import "time"

// Background notifications only enqueue work; they never mutate the network.
func (e *Engine) requestTunnelRecovery() {
	if e.context.Err() == nil {
		e.queue.Add(e.recoveryRequest)
	}
}

// The driver selects the delay. Engine's configuration-event rate limiter is
// unchanged, and this queue item is consumed by the same reconciliation worker.
func (e *Engine) scheduleDriver() {
	if e.queue == nil || e.context.Err() != nil {
		return
	}
	if delay := e.nextDriverReconcile(); delay > 0 {
		e.queue.AddAfter(e.recoveryRequest, delay)
	}
}

func (e *Engine) nextDriverReconcile() time.Duration {
	if driver, ok := e.tunnel.vpnDriver.(interface{ NextReconcile() time.Duration }); ok {
		return driver.NextReconcile()
	}
	return 0
}
