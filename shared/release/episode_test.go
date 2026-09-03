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
		{"House.of.the.Dragon.S01E01-E03.1080p.WEB-DL", 1, 2, true},
		{"House.of.the.Dragon.S01E01-E03.1080p.WEB-DL", 1, 4, false},
		{"Doctor.Who.2005.8x11.Dark.Water.720p.HDTV", 8, 11, true},
		{"Show.S00E02.Special.1080p.WEB-DL", 0, 2, true},
		{"The Ed Show 10-19-12.mp4", 10, 19, false},
	}
	for _, c := range cases {
		if got := MatchesEpisode(c.title, c.season, c.ep); got != c.want {
			t.Errorf("MatchesEpisode(%q, %d, %d) = %v, want %v", c.title, c.season, c.ep, got, c.want)
		}
	}
}
