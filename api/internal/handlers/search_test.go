package handlers

import "testing"

func TestFileMatchesEpisode(t *testing.T) {
	if !fileMatchesEpisode("Severance.S01E03.1080p.mkv", 1, 3) {
		t.Error("S01E03 should match s1e3")
	}
	if fileMatchesEpisode("Severance.S01E03.1080p.mkv", 2, 3) {
		t.Error("S01E03 should not match s2e3")
	}
}

func TestYearMatches(t *testing.T) {
	if !yearMatches("The.Matrix.1999.1080p.mkv", 1999) {
		t.Error("1999 release should match year 1999")
	}
	if yearMatches("The.Matrix.1999.1080p.mkv", 2003) {
		t.Error("1999 release should not match year 2003")
	}
	if !yearMatches("Some.Yearless.Release.mkv", 2010) {
		t.Error("yearless release should be lenient (match)")
	}
}
