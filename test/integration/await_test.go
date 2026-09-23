package main

import (
	"testing"
	"time"
)

// wait for condition to be true
func await(t *testing.T, interval time.Duration, timeout time.Duration, f func() bool) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if f() {
			return
		}
		time.Sleep(interval)
	}
	t.Errorf("Timeout!")
}

func TestAwaitSucceedsOnConditionMet(t *testing.T) {
	calls := 0
	await(t, 10*time.Millisecond, time.Second, func() bool {
		calls++
		return calls >= 3
	})
	if calls < 3 {
		t.Errorf("expected at least 3 calls, got %d", calls)
	}
}

func TestAwaitTimesOutCleanly(t *testing.T) {
	mockT := &testing.T{}
	start := time.Now()
	await(mockT, 10*time.Millisecond, 50*time.Millisecond, func() bool {
		return false
	})
	elapsed := time.Since(start)
	if elapsed > 200*time.Millisecond {
		t.Errorf("await took too long to time out: %v", elapsed)
	}
}
