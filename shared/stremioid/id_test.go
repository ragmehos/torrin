package stremioid

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want ID
	}{
		{name: "movie", raw: "tt1234567", want: ID{IMDBID: "1234567"}},
		{name: "episode", raw: "tt1234567:5:1", want: ID{IMDBID: "1234567", Season: 5, Episode: 1}},
		{name: "uppercase hash", raw: "C12FE1C06BBA254A9DC9F519B335AA7C1367A88A", want: ID{InfoHash: "c12fe1c06bba254a9dc9f519b335aa7c1367a88a"}},
		{name: "bad episode", raw: "tt1234567:five:1", want: ID{IMDBID: "1234567"}},
		{name: "extra episode part", raw: "tt1234567:5:1:extra", want: ID{IMDBID: "1234567"}},
		{name: "garbage", raw: "garbage", want: ID{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Parse(tt.raw); got != tt.want {
				t.Fatalf("Parse(%q) = %+v, want %+v", tt.raw, got, tt.want)
			}
		})
	}
}

func TestIsEpisode(t *testing.T) {
	if !Parse("tt1234567:5:1").IsEpisode() {
		t.Fatal("valid series ID should be an episode")
	}
	if Parse("tt1234567").IsEpisode() {
		t.Fatal("movie ID should not be an episode")
	}
}
