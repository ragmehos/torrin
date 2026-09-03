package stremioid

import "testing"

func TestParse(t *testing.T) {
	hash := "c12fe1c06bba254a9dc9f519b335aa7c1367a88a"
	tests := []struct {
		raw             string
		imdb, infoHash  string
		season, episode int
		isEpisode       bool
	}{
		{raw: "tt1234567", imdb: "1234567"},
		{raw: "tt1234567:12:3", imdb: "1234567", season: 12, episode: 3, isEpisode: true},
		{raw: "tt1234567:0:2", imdb: "1234567", episode: 2, isEpisode: true},
		{raw: hash, infoHash: hash},
		{raw: "tt1234567:bad:3"},
		{raw: "tt1234567:12:0"},
		{raw: "garbage"},
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			got := Parse(tt.raw)
			if got.IMDBID != tt.imdb || got.InfoHash != tt.infoHash || got.Season != tt.season || got.Episode != tt.episode || got.IsEpisode() != tt.isEpisode {
				t.Fatalf("Parse(%q) = %+v", tt.raw, got)
			}
		})
	}
}
