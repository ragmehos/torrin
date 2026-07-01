package providers

import "testing"

// reapDecision mirrors ADReapOrphans' per-magnet rule: terminal-error magnets
// (statusCode >= 5) and stale "ready" cache-checks (code 4, older than 20 min)
// get deleted; processing magnets (0-3) and fresh ready ones are left alone.
func reapDecision(statusCode int, ageSecs int64) bool {
	dead := statusCode >= 5
	readyStale := statusCode == 4 && ageSecs > 20*60
	return dead || readyStale
}

func TestReapDecision(t *testing.T) {
	cases := []struct {
		name       string
		statusCode int
		ageSecs    int64
		wantReap   bool
	}{
		{"in queue", 0, 10, false},
		{"downloading", 1, 10, false},
		{"uploading", 3, 10, false},
		{"fresh ready", 4, 60, false},
		{"stale ready cache-check", 4, 25 * 60, true},
		{"upload fail", 5, 10, true},
		{"file too big", 8, 10, true},       // string-match used to miss this
		{"deleted on hoster", 11, 10, true}, // and this
		{"processing failed", 12, 10, true}, // and this
		{"no peer", 15, 10, true},
	}
	for _, c := range cases {
		if got := reapDecision(c.statusCode, c.ageSecs); got != c.wantReap {
			t.Errorf("%s (code %d, age %ds): reap=%v want %v", c.name, c.statusCode, c.ageSecs, got, c.wantReap)
		}
	}
}
