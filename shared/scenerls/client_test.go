package scenerls

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/torrin-app/torrin/shared/release"
)

func TestIsPost(t *testing.T) {
	cases := map[string]bool{
		"https://scene-rls.net/dune-part-two-2024-1080p-bluray-ddp7-1-x264-mhysa/": true,
		"https://scene-rls.net/releases/":                                          false,
		"https://scene-rls.net/contact/":                                           false,
		"https://scene-rls.net/category/movies/":                                   false,
		"https://scene-rls.net/tag/1080p/":                                         false,
		"https://scene-rls.net/?s=Dune":                                            false,
		"https://other.com/dune-part-two-2024/":                                    false,
		"https://scene-rls.net/short/":                                             false,
	}
	for href, want := range cases {
		if got := isPost(href); got != want {
			t.Errorf("isPost(%q)=%v want %v", href, got, want)
		}
	}
}

func TestDotted(t *testing.T) {
	if got := dotted("Frieren: Beyond Journey's End"); got != "frieren.beyond.journeys.end" {
		t.Errorf("dotted = %q", got)
	}
}

func TestLargestSize(t *testing.T) {
	if got := largestSize("sample 50 MB main 10.46 GB proof 5 MB"); got != "10.46 GB" {
		t.Errorf("largestSize = %q", got)
	}
	if got := largestSize("700 mb file"); got != "700 MB" {
		t.Errorf("largestSize = %q", got)
	}
	if got := largestSize("no size here"); got != "" {
		t.Errorf("largestSize = %q", got)
	}
}

func TestIMDBFromDoc(t *testing.T) {
	html := `<html><body>
		<a href="https://scene-rls.net/?s=IMDB+Top">IMDB Top</a>
		<a href="https://www.imdb.com/title/tt15239678/">IMDB</a>
	</body></html>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	if got := imdbFromDoc(doc); got != "tt15239678" {
		t.Errorf("imdbFromDoc = %q", got)
	}
	none, _ := goquery.NewDocumentFromReader(strings.NewReader(`<a href="https://x.com">x</a>`))
	if got := imdbFromDoc(none); got != "" {
		t.Errorf("imdbFromDoc(none) = %q", got)
	}
}

func TestFilter(t *testing.T) {
	rs := []release.Result{
		{Title: "Dune Part Two 2024 1080p"},
		{Title: "Dune Part One 2021 1080p"},
		{Title: "Show S01E05 1080p"},
		{Title: "Show S01E06 1080p"},
	}
	if got := filter(rs, "2024", ""); len(got) != 1 || !strings.Contains(got[0].Title, "2024") {
		t.Errorf("year filter = %+v", got)
	}
	if got := filter(rs, "", "s01e05"); len(got) != 1 || !strings.Contains(got[0].Title, "S01E05") {
		t.Errorf("ep filter = %+v", got)
	}
	if got := filter(rs, "", ""); len(got) != 4 {
		t.Error("empty filter should return all")
	}
}

func TestOrderedViews(t *testing.T) {
	html := `<html><body>
		<a href="http://nfo.scene-rls.net/view/111">NitroFlare</a>
		<a href="http://nfo.scene-rls.net/view/222">RapidGator</a>
		<a href="https://scene-rls.net/other/">other</a>
	</body></html>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	got := orderedViews(doc)
	if len(got) != 2 {
		t.Fatalf("got %d views", len(got))
	}
	if !strings.HasSuffix(got[0], "/222") {
		t.Errorf("rapidgator should be first, got %v", got)
	}
}
