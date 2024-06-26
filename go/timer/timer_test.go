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

package timer

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	half    = 50 * time.Millisecond
	quarter = 25 * time.Millisecond
	tenth   = 10 * time.Millisecond
)

var numcalls atomic.Int64

func f() {
	numcalls.Add(1)
}

func TestWait(t *testing.T) {
	numcalls.Store(0)
	timer := NewTimer(quarter)
	assert.False(t, timer.Running())
	timer.Start(f)
	defer timer.Stop()
	assert.True(t, timer.Running())
	time.Sleep(tenth)
	assert.Equal(t, int64(0), numcalls.Load())
	time.Sleep(quarter)
	assert.Equal(t, int64(1), numcalls.Load())
	time.Sleep(quarter)
	assert.Equal(t, int64(2), numcalls.Load())
}

func TestReset(t *testing.T) {
	numcalls.Store(0)
	timer := NewTimer(half)
	timer.Start(f)
	defer timer.Stop()
	timer.SetInterval(quarter)
	time.Sleep(tenth)
	assert.Equal(t, int64(0), numcalls.Load())
	time.Sleep(quarter)
	assert.Equal(t, int64(1), numcalls.Load())
}

func TestIndefinite(t *testing.T) {
	numcalls.Store(0)
	timer := NewTimer(0)
	timer.Start(f)
	defer timer.Stop()
	timer.TriggerAfter(quarter)
	time.Sleep(tenth)
	assert.Equal(t, int64(0), numcalls.Load())
	time.Sleep(quarter)
	assert.Equal(t, int64(1), numcalls.Load())
}

func TestInterval(t *testing.T) {
	timer := NewTimer(100)
	in := timer.Interval()
	assert.Equal(t, 100*time.Nanosecond, in)

	timer.interval.Store(200)
	in = timer.Interval()
	assert.Equal(t, 200*time.Nanosecond, in)
}
