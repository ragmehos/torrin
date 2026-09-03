package addon

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/torrin-app/torrin/shared/jobs"
	"github.com/torrin-app/torrin/shared/stremioid"
)

func TestStreamTargetMatchesType(t *testing.T) {
	tests := []struct {
		contentType, contentID string
		want                   bool
	}{
		{"movie", "tt1234567", true},
		{"series", "tt1234567:12:2", true},
		{"series", "tt1234567", false},
		{"movie", "tt1234567:12:2", false},
		{"series", "tt1234567:bad:2", false},
	}
	for _, tt := range tests {
		if got := streamTargetMatchesType(tt.contentType, stremioid.Parse(tt.contentID)); got != tt.want {
			t.Errorf("%s/%s = %v, want %v", tt.contentType, tt.contentID, got, tt.want)
		}
	}
}

func TestLibraryFilesUseRequestedSeasonNotStoredSeason(t *testing.T) {
	j := &jobs.Job{Season: 12, Episode: 1, Files: []jobs.File{
		{Name: "Paw.Patrol.S05E01.mkv"},
		{Name: "Paw.Patrol.S12E01.mkv"},
	}}
	got := jobs.FilesForEpisode(j, j.Files, 5, 1)
	if len(got) != 1 || got[0].Name != "Paw.Patrol.S05E01.mkv" {
		t.Fatalf("files = %+v", got)
	}
}

func TestEntryUsesStremioStreamFields(t *testing.T) {
	hash := "c12fe1c06bba254a9dc9f519b335aa7c1367a88a"
	got := entry(`Show\Season 12\Show.S12E03.mkv`, "https://stream.example/file", hash, 1234)
	if got["title"] != "Show.S12E03.mkv" || got["description"] != "Show.S12E03.mkv" {
		t.Fatalf("display fields = %+v", got)
	}
	hints := got["behaviorHints"].(map[string]any)
	if hints["filename"] != "Show.S12E03.mkv" || hints["videoSize"] != int64(1234) || hints["bingeGroup"] != "torrin:"+hash {
		t.Fatalf("behavior hints = %+v", hints)
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
