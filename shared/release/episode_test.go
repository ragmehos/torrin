package release

import "testing"

func TestMatchesEpisode(t *testing.T) {
	cases := []struct {
		title      string
		season, ep int
		want       bool
	}{
		{"House.of.the.Dragon.S01E01.1080p.WEB-DL", 1, 1, true},
		{"House.of.the.Dragon.S01E05.1080p.WEB-DL", 1, 1, false},
		{"The.Last.of.Us.S01.2160p.WEB-DL.x265", 1, 1, true},  // season pack covers ep 1
		{"The.Last.of.Us.S02.1080p.BluRay.x264", 1, 1, false}, // wrong season pack
		{"House.of.the.Dragon.S01E01.1080p.WEB-DL", 2, 1, false},
	}
	for _, c := range cases {
		if got := MatchesEpisode(c.title, c.season, c.ep); got != c.want {
			t.Errorf("MatchesEpisode(%q, %d, %d) = %v, want %v", c.title, c.season, c.ep, got, c.want)
		}
	}
}
