package addon

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/torrin-app/torrin/shared/jobs"
	"github.com/torrin-app/torrin/shared/stremioid"
)

func TestLibraryFilesFiltersRequestedEpisode(t *testing.T) {
	j := &jobs.Job{Season: 5, Files: []jobs.File{
		{Name: "Paw.Patrol.S05E01.mkv"},
		{Name: "Paw.Patrol.S05E02.mkv"},
		{Name: "Paw.Patrol.S12E01.mkv"},
	}}
	got := libraryFiles(j, stremioid.Parse("tt3121722:5:1"))
	if len(got) != 1 {
		t.Fatalf("files = %d, want one S05E01 file", len(got))
	}
	if got[0].Name != "Paw.Patrol.S05E01.mkv" || got[0].Index != 0 {
		t.Fatalf("file = %+v, want the original S05E01 file", got[0])
	}
}

func TestLibraryFilesDoesNotReturnWrongSeasonPack(t *testing.T) {
	j := &jobs.Job{Season: 12, Files: []jobs.File{{Name: "Paw.Patrol.S12E01.mkv"}}}
	if got := libraryFiles(j, stremioid.Parse("tt3121722:5:1")); len(got) != 0 {
		t.Fatalf("wrong-season files = %v, want none", got)
	}
}

func TestLibraryFilesDoesNotFilterMovies(t *testing.T) {
	j := &jobs.Job{Files: []jobs.File{
		{Name: "Movie.2024.mkv"},
		{Name: "Movie.behind-the-scenes.mkv"},
	}}
	got := libraryFiles(j, stremioid.Parse("tt0816692"))
	if len(got) != len(j.Files) {
		t.Fatalf("movie files = %d, want %d", len(got), len(j.Files))
	}
}

func TestStreamTargetMatchesType(t *testing.T) {
	tests := []struct {
		contentType string
		contentID   string
		want        bool
	}{
		{contentType: "movie", contentID: "tt0816692", want: true},
		{contentType: "series", contentID: "tt3121722:5:1", want: true},
		{contentType: "series", contentID: "tt3121722:0:2", want: true},
		{contentType: "series", contentID: "tt3121722"},
		{contentType: "movie", contentID: "tt3121722:5:1"},
		{contentType: "series", contentID: "tt3121722:bad:1"},
		{contentType: "other", contentID: "tt0816692"},
	}
	for _, tt := range tests {
		t.Run(tt.contentType+"/"+tt.contentID, func(t *testing.T) {
			if got := streamTargetMatchesType(tt.contentType, stremioid.Parse(tt.contentID)); got != tt.want {
				t.Fatalf("streamTargetMatchesType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEntryIncludesPlaybackHints(t *testing.T) {
	hash := "C12FE1C06BBA254A9DC9F519B335AA7C1367A88A"
	got := entry(`Show\Season 05\Show.S05E01.MKV`, "https://stream.example/file", hash, 1234)
	if got["title"] != "Show.S05E01.MKV" {
		t.Fatalf("title = %v, want base filename", got["title"])
	}
	if got["description"] != "Show.S05E01.MKV" {
		t.Fatalf("description = %v, want base filename", got["description"])
	}
	hints := got["behaviorHints"].(map[string]any)
	if hints["filename"] != "Show.S05E01.MKV" {
		t.Errorf("filename = %v", hints["filename"])
	}
	if hints["notWebReady"] != true {
		t.Errorf("notWebReady = %v, want true for MKV", hints["notWebReady"])
	}
	if hints["bingeGroup"] != "torrin:c12fe1c06bba254a9dc9f519b335aa7c1367a88a" {
		t.Errorf("bingeGroup = %v", hints["bingeGroup"])
	}
	if hints["videoSize"] != int64(1234) {
		t.Errorf("videoSize = %v", hints["videoSize"])
	}

	mp4 := entry("movie.MP4", "https://stream.example/file", hash, 0)
	mp4Hints := mp4["behaviorHints"].(map[string]any)
	if mp4Hints["notWebReady"] != false {
		t.Errorf("HTTPS MP4 notWebReady = %v, want false", mp4Hints["notWebReady"])
	}
	if _, ok := mp4Hints["videoSize"]; ok {
		t.Error("unknown video size should be omitted")
	}
}

func TestManifest(t *testing.T) {
	w := httptest.NewRecorder()
	(&Server{}).manifest(w, httptest.NewRequest(http.MethodGet, "/key/manifest.json", nil))
	if w.Code != 200 {
		t.Fatalf("status %d", w.Code)
	}
	var m map[string]any
	json.Unmarshal(w.Body.Bytes(), &m)
	if m["id"] != "app.torrin.stremio" {
		t.Errorf("manifest id = %v", m["id"])
	}
}
