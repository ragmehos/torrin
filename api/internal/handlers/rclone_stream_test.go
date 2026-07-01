package handlers

import "testing"

func TestRcloneStreamSig(t *testing.T) {
	key := []byte("signing-key")
	a := rcloneStreamSig(key, "job1", 0, 1750000000)
	if a == "" || len(a) != 64 { // hex sha256
		t.Fatalf("bad sig: %q", a)
	}
	if rcloneStreamSig(key, "job1", 0, 1750000000) != a {
		t.Error("not deterministic")
	}
	if rcloneStreamSig(key, "job1", 1, 1750000000) == a {
		t.Error("idx should change sig")
	}
	if rcloneStreamSig([]byte("other"), "job1", 0, 1750000000) == a {
		t.Error("key should change sig")
	}
}
