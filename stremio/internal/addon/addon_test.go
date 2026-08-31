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
