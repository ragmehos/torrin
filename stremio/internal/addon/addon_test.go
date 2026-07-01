package addon

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseID(t *testing.T) {
	if imdb, hash := parseID("tt1234567"); imdb != "1234567" || hash != "" {
		t.Errorf("movie: imdb=%q hash=%q", imdb, hash)
	}
	if imdb, _ := parseID("tt1234567:1:2"); imdb != "1234567" {
		t.Errorf("series imdb=%q", imdb)
	}
	h := "c12fe1c06bba254a9dc9f519b335aa7c1367a88a"
	if imdb, hash := parseID(h); hash != h || imdb != "" {
		t.Errorf("hash: imdb=%q hash=%q", imdb, hash)
	}
	if imdb, hash := parseID("garbage"); imdb != "" || hash != "" {
		t.Errorf("garbage should be empty, got %q %q", imdb, hash)
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
