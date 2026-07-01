package handlers

import "testing"

func TestGenLinkCode(t *testing.T) {
	const allowed = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		c := genLinkCode()
		if len(c) != 8 {
			t.Fatalf("len = %d", len(c))
		}
		for _, ch := range c {
			if !contains(allowed, ch) {
				t.Fatalf("ambiguous char %q in %q", ch, c)
			}
		}
		seen[c] = true
	}
	if len(seen) < 45 {
		t.Errorf("low entropy: %d unique of 50", len(seen))
	}
}

func contains(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}
