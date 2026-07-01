package jobs

import (
	"testing"
	"time"
)

func TestCalcStreak(t *testing.T) {
	d := func(s string) time.Time { tm, _ := time.Parse("2006-01-02", s); return tm }
	if got := calcStreak(nil); got != 0 {
		t.Errorf("empty = %d want 0", got)
	}
	// Consecutive days (newest first) → streak of 3.
	if got := calcStreak([]time.Time{d("2026-06-24"), d("2026-06-23"), d("2026-06-22")}); got != 3 {
		t.Errorf("3-day streak = %d", got)
	}
	// Gap after first two breaks the streak at 2.
	if got := calcStreak([]time.Time{d("2026-06-24"), d("2026-06-23"), d("2026-06-20")}); got != 2 {
		t.Errorf("broken streak = %d want 2", got)
	}
}
