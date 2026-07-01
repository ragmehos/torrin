package handlers

import (
	"testing"
	"time"
)

func TestOAuthState(t *testing.T) {
	key := []byte("k")
	st := signOAuthState(key, "user1", "dropbox", time.Now().Add(time.Minute).Unix())
	uid, prov, ok := verifyOAuthState(key, st)
	if !ok || uid != "user1" || prov != "dropbox" {
		t.Fatalf("roundtrip failed: %q %q %v", uid, prov, ok)
	}
	if _, _, ok := verifyOAuthState([]byte("wrong"), st); ok {
		t.Error("tampered key should fail")
	}
	expired := signOAuthState(key, "u", "dropbox", time.Now().Add(-time.Minute).Unix())
	if _, _, ok := verifyOAuthState(key, expired); ok {
		t.Error("expired state should fail")
	}
}
