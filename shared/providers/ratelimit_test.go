package providers

import (
	"context"
	"testing"
	"time"
)

func TestSharedLimitersAreSharedPerKey(t *testing.T) {
	a := sharedLimiters("test:k1", limitSpec{5, time.Minute})
	b := sharedLimiters("test:k1", limitSpec{5, time.Minute})
	if a != b {
		t.Fatal("same id must return the same shared limiter instance")
	}
	if sharedLimiters("test:k2", limitSpec{5, time.Minute}) == a {
		t.Fatal("different id must return a different limiter")
	}
}

func TestSharedLimitersMultipleBuckets(t *testing.T) {
	k := sharedLimiters("test:dual", limitSpec{12, time.Second}, limitSpec{600, time.Minute})
	if len(k.specs) != 2 || len(k.mem) != 2 {
		t.Fatalf("expected 2 buckets, got %d", len(k.specs))
	}
}

func TestWaitLimitsThrottlesWhenEmpty(t *testing.T) {
	k := sharedLimiters("test:throttle", limitSpec{2, time.Minute})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	waitLimits(ctx, k)
	waitLimits(ctx, k)

	done := make(chan struct{})
	go func() {
		waitLimits(ctx, k)
		close(done)
	}()
	select {
	case <-done:
		t.Fatal("third call should block on an empty bucket")
	case <-time.After(50 * time.Millisecond):
	}
}
