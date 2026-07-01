package providers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHosterUnlock(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v4/link/unlock" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if r.URL.Query().Get("link") != "https://host/file" {
			t.Errorf("link not forwarded: %s", r.URL.RawQuery)
		}
		w.Write([]byte(`{"status":"success","data":{"link":"https://dl/x.mkv","filename":"x.mkv","filesize":123}}`))
	}))
	defer srv.Close()

	old := adBase
	adBase = srv.URL
	defer func() { adBase = old }()

	name, dl, size, err := HosterUnlock(context.Background(), "key", "https://host/file")
	if err != nil {
		t.Fatal(err)
	}
	if name != "x.mkv" || dl != "https://dl/x.mkv" || size != 123 {
		t.Fatalf("got %q %q %d", name, dl, size)
	}
}

func TestHosterUnlockError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"error","error":{"code":"LINK_HOST_NOT_SUPPORTED","message":"nope"}}`))
	}))
	defer srv.Close()

	old := adBase
	adBase = srv.URL
	defer func() { adBase = old }()

	if _, _, _, err := HosterUnlock(context.Background(), "key", "https://host/file"); err == nil {
		t.Fatal("expected error from failed unlock")
	}
}
