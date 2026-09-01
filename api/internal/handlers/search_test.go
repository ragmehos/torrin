package handlers

import (
	"testing"

	"github.com/torrin-app/torrin/shared/jobs"
)

func TestEpisodeMatch(t *testing.T) {
	tagged := &jobs.Job{Season: 2, Episode: 5, Files: []jobs.File{{Name: "whatever_garbage_name.mkv"}}}
	if !episodeMatch(tagged, "whatever_garbage_name.mkv", 2, 5) {
		t.Error("single-file tagged job should fall back to stored season/episode for an opaque filename")
	}
	if episodeMatch(tagged, "whatever_garbage_name.mkv", 1, 5) {
		t.Error("tagged job for s2e5 must not match s1e5")
	}
	pack := &jobs.Job{Season: 2, Episode: 5, Files: []jobs.File{{Name: "opaque_a.mkv"}, {Name: "opaque_b.mkv"}}}
	if episodeMatch(pack, "opaque_a.mkv", 2, 5) {
		t.Error("opaque file in a multi-file pack must not match on shared job metadata")
	}
	untagged := &jobs.Job{}
	if !episodeMatch(untagged, "Severance.S01E03.1080p.mkv", 1, 3) {
		t.Error("untagged job falls back to filename parse")
	}
	if episodeMatch(untagged, "no_episode_here.mkv", 1, 3) {
		t.Error("untagged job with unparseable name must not match")
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
