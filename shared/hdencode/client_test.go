package hdencode

import "testing"

func TestIsPost(t *testing.T) {
	cases := map[string]bool{
		"https://hdencode.org/dune-part-two-2024-1080p-bluray-h264-risehd-30-5-gb/": true,
		"https://hdencode.org/advanced-search/":                                     false,
		"https://hdencode.org/watchlist/":                                           false,
		"https://hdencode.org/category/movies/":                                     false,
		"https://hdencode.org/tag/1080p/":                                           false,
		"https://other.com/dune-part-two-2024-1080p/":                               false,
		"https://hdencode.org/short/":                                               false, // too short
	}
	for href, want := range cases {
		if got := isPost(href); got != want {
			t.Errorf("isPost(%q)=%v want %v", href, got, want)
		}
	}
}

func TestFilterEp(t *testing.T) {
	rs := []Result{
		{Title: "Show.S01E05.1080p.WEB-DL"},
		{Title: "Show.S01E06.1080p.WEB-DL"},
		{Title: "Show.S02E05.1080p.WEB-DL"},
		{Title: "Show.S01.1080p.BluRay.x264-GRP"},
	}
	got := filterEp(rs, 1, 5)
	if len(got) != 2 {
		t.Fatalf("want episode + season pack (2), got %d: %+v", len(got), got)
	}
	if len(filterEp(rs, 0, 0)) != 4 {
		t.Error("movie (season 0) should return all")
	}
}
