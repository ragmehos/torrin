package bot

import "testing"

func TestDocCacheKey(t *testing.T) {
	// Deterministic 40-hex (sha1) per Telegram doc id — the dedup key.
	a := docCacheKey(12345)
	if len(a) != 40 {
		t.Fatalf("want 40-hex, got %d chars", len(a))
	}
	if docCacheKey(12345) != a {
		t.Error("not deterministic")
	}
	if docCacheKey(54321) == a {
		t.Error("different ids collided")
	}
}

func TestUserLimiter(t *testing.T) {
	l := newUserLimiter()
	first := l.tryAcquire("u", 2)
	second := l.tryAcquire("u", 2)
	if !first || !second {
		t.Fatal("should allow 2")
	}
	if l.tryAcquire("u", 2) {
		t.Error("3rd should be denied")
	}
	l.release("u")
	if !l.tryAcquire("u", 2) {
		t.Error("should allow after release")
	}
}
