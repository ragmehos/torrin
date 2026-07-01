package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestXMLEscape(t *testing.T) {
	if got := xmlEscape(`a&b<c>"d"`); got != `a&amp;b&lt;c&gt;&quot;d&quot;` {
		t.Errorf("got %q", got)
	}
}

func TestOptions(t *testing.T) {
	w := httptest.NewRecorder()
	(&Server{}).serve(w, httptest.NewRequest(http.MethodOptions, "/", nil))
	if w.Code != 200 || w.Header().Get("DAV") != "1, 2" {
		t.Errorf("OPTIONS: code=%d dav=%q", w.Code, w.Header().Get("DAV"))
	}
}

func TestUnauthorized(t *testing.T) {
	// No Basic auth → 401 before any store access (so nil deps are fine).
	w := httptest.NewRecorder()
	(&Server{}).serve(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Code != 401 || w.Header().Get("WWW-Authenticate") == "" {
		t.Errorf("expected 401 challenge, got %d", w.Code)
	}
}
