package auth

import (
	"testing"
	"time"
)

func TestIsPaused(t *testing.T) {
	var u User
	if u.IsPaused() {
		t.Error("zero PausedAt should not count as paused")
	}
	u.PausedAt = time.Now()
	if !u.IsPaused() {
		t.Error("set PausedAt should count as paused")
	}
}
