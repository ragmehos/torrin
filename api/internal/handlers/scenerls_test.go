package handlers

import (
	"strings"
	"testing"

	"github.com/torrin-app/torrin/shared/release"
)

func TestMatchReleaseTitle(t *testing.T) {
	results := []release.Result{
		{Title: "House of the Dragon S01E01 The Heirs of the Dragon 2160p MAX WEB-DL TrueHD 7 1 Atmos DV HDR H265-Kitsune"},
		{Title: "House of the Dragon S01E01 1080p WEB H264-CAKES"},
		{Title: "The House That Dragons Built S01E01 720p WEB h264-KOGi"},
		{Title: "House That Dragons Built S01E01-E09 1080p HMAX WEB-DL AAC2 0 x265-Slartibartfast"},
	}
	got := matchReleaseTitle([]string{"House of the Dragon"}, results)
	if len(got) != 2 {
		t.Fatalf("want 2 kept, got %d: %+v", len(got), got)
	}
	for _, r := range got {
		if !strings.HasPrefix(r.Title, "House of the Dragon S01E01") {
			t.Errorf("unexpected kept: %q", r.Title)
		}
	}
	if len(matchReleaseTitle([]string{""}, results)) != len(results) {
		t.Error("empty want should keep all")
	}
	// an alias also matches
	if len(matchReleaseTitle([]string{"Wrong", "House of the Dragon"}, results)) != 2 {
		t.Error("alias should match")
	}
}
