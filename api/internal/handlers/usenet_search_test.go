package handlers

import "testing"

func TestNormalizeIndexerURL(t *testing.T) {
	cases := map[string]string{
		"https://idx.com/":     "https://idx.com",
		"https://idx.com/api":  "https://idx.com",
		"https://idx.com/api/": "https://idx.com",
		"  https://idx.com  ":  "https://idx.com",
	}
	for in, want := range cases {
		if got := normalizeIndexerURL(in); got != want {
			t.Errorf("normalizeIndexerURL(%q)=%q want %q", in, got, want)
		}
	}
}
