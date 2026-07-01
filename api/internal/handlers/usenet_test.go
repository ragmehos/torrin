package handlers

import (
	"testing"

	"github.com/torrin-app/torrin/shared/plans"
)

func TestUsenetEntitled(t *testing.T) {
	sys := plans.Plan{SystemUsenet: true}
	noSys := plans.Plan{SystemUsenet: false}
	cases := []struct {
		name     string
		ownCreds bool
		plan     plans.Plan
		want     bool
	}{
		{"own creds, no system plan", true, noSys, true},
		{"own creds, system plan", true, sys, true},
		{"no creds, system plan", false, sys, true},
		{"no creds, no system plan", false, noSys, false}, // the newly-blocked case
	}
	for _, c := range cases {
		if got := usenetEntitled(c.ownCreds, c.plan); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
