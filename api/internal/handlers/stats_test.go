package handlers

import (
	"testing"
	"time"
)

func TestWrappedAvailable(t *testing.T) {
	// June has 30 days → window opens on the 24th (lastDay-6).
	if wrappedAvailable(time.Date(2026, 6, 23, 0, 0, 0, 0, time.UTC)) {
		t.Error("June 23 should be closed")
	}
	if !wrappedAvailable(time.Date(2026, 6, 24, 0, 0, 0, 0, time.UTC)) {
		t.Error("June 24 should be open")
	}
	if !wrappedAvailable(time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)) {
		t.Error("June 30 should be open")
	}
}
