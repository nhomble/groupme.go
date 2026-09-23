package main

import (
	"testing"
	"time"
)

// wait for condition to be true
func await(t *testing.T, interval time.Duration, timeout time.Duration, f func() bool) {
	complete := make(chan bool)

	go func() {
		for !f() {
			time.Sleep(interval)
		}
		complete <- true
	}()

	select {
	case <-complete:
	case <-time.After(timeout):
		t.Errorf("Timeout!")
	}
}
