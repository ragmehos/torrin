package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) })
}

func req(key, cfip string) *http.Request {
	r := httptest.NewRequest("GET", "/api/jobs", nil)
	if key != "" {
		r.Header.Set("Authorization", "Bearer "+key)
	}
	if cfip != "" {
		r.Header.Set("CF-Connecting-IP", cfip)
	}
	return r
}

func codeFor(h http.Handler, r *http.Request) int {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w.Code
}

func TestRateLimitPerKey(t *testing.T) {
	h := RateLimit(1, 2, nil)(okHandler())
	// burst of 2 passes, then throttled
	for i := 1; i <= 2; i++ {
		if got := codeFor(h, req("user", "1.2.3.4")); got != 200 {
			t.Fatalf("burst req %d got %d, want 200", i, got)
		}
	}
	if got := codeFor(h, req("user", "1.2.3.4")); got != 429 {
		t.Fatalf("over-burst got %d, want 429", got)
	}
	// a different key is independent
	if got := codeFor(h, req("other", "1.2.3.4")); got != 200 {
		t.Fatalf("different key got %d, want 200", got)
	}
}

func TestRateLimitExempt(t *testing.T) {
	h := RateLimit(1, 1, []string{"servicekey"})(okHandler())
	for i := 0; i < 5; i++ {
		if codeFor(h, req("servicekey", "1.2.3.4")) != 200 {
			t.Fatalf("exempt key throttled on req %d", i)
		}
	}
}

func TestRateLimitInternalBypass(t *testing.T) {
	h := RateLimit(1, 1, nil)(okHandler())
	for i := 0; i < 5; i++ {
		if codeFor(h, req("", "")) != 200 { // no key, no CF header
			t.Fatalf("internal call throttled on req %d", i)
		}
	}
}

func TestRateLimitDisabled(t *testing.T) {
	h := RateLimit(0, 0, nil)(okHandler())
	for i := 0; i < 5; i++ {
		if codeFor(h, req("k", "1.2.3.4")) != 200 {
			t.Fatalf("disabled limiter throttled on req %d", i)
		}
	}
}
