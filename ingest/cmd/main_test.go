package main

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestTrackIfAbsent(t *testing.T) {
	c := newCancelRegistry()
	noop := func() {}
	if !c.trackIfAbsent("a", noop) {
		t.Fatal("first track should win")
	}
	if c.trackIfAbsent("a", noop) {
		t.Fatal("duplicate track must be rejected")
	}
	c.untrack("a")
	if !c.trackIfAbsent("a", noop) {
		t.Fatal("track after untrack should win")
	}
}

func TestTrackIfAbsentConcurrent(t *testing.T) {
	c := newCancelRegistry()
	var wins int64
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if c.trackIfAbsent("x", func() {}) {
				atomic.AddInt64(&wins, 1)
			}
		}()
	}
	wg.Wait()
	if wins != 1 {
		t.Fatalf("expected exactly 1 winner under concurrency, got %d", wins)
	}
}
