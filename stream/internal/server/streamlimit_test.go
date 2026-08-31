package server

import (
	"net/http/httptest"
	"testing"
)

func TestStreamGuard(t *testing.T) {
	g := newStreamGuard(2, 1)
	if got := g.acquire("u1"); got != streamAllowed {
		t.Fatalf("first user acquire = %v", got)
	}
	if got := g.acquire("u1"); got != streamLimitUser {
		t.Fatalf("second user acquire = %v", got)
	}
	if got := g.acquire("u2"); got != streamAllowed {
		t.Fatalf("other user acquire = %v", got)
	}
	if got := g.acquire("u3"); got != streamLimitTotal {
		t.Fatalf("global acquire = %v", got)
	}
	g.release("u1")
	if got := g.acquire("u3"); got != streamAllowed {
		t.Fatalf("acquire after release = %v", got)
	}
	g.release("missing")
}

func TestCairnStreamIdentity(t *testing.T) {
	userReq := httptest.NewRequest("GET", "/file?u=user-1&sig=abc", nil)
	if got := cairnStreamIdentity(userReq); got != "user:user-1" {
		t.Fatalf("user identity = %q", got)
	}
	legacyReq := httptest.NewRequest("GET", "/file?sig=abc", nil)
	if got := cairnStreamIdentity(legacyReq); got != "legacy:abc" {
		t.Fatalf("legacy identity = %q", got)
	}
}
