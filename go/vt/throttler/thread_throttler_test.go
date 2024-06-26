/*
Copyright 2019 The Vitess Authors.

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

package throttler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestThrottle_NoBurst(t *testing.T) {
	tt := newThreadThrottler(0, newAggregatedIntervalHistory(1, 1*time.Second, 1))
	tt.setMaxRate(2)
	// We set the rate to 2 requests per second, and internally the throttler uses a burst value of
	// 1. This means that in any time interval of length t seconds, the throttler should
	// not allow more than floor(2*t+1) requests. For example, in the interval [1500ms, 1501ms], of
	// length 1ms, we shouldn't be able to send more than floor(2*10^-3+1)=1 requests.
	gotBackoff := tt.throttle(sinceZero(1500 * time.Millisecond))
	require.Equal(t, NotThrottled, gotBackoff, "throttler should not have throttled us")

	wantBackoff := 499 * time.Millisecond
	gotBackoff = tt.throttle(sinceZero(1501 * time.Millisecond))
	require.Equal(t, wantBackoff, gotBackoff, "throttler should have throttled us")
}
