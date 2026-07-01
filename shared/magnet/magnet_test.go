package magnet

import (
	"encoding/base32"
	"encoding/hex"
	"testing"
)

func TestHashHexAndBase32(t *testing.T) {
	want := "5a181a17d765ee945c1786503bd8bdba67085226"
	raw, _ := hex.DecodeString(want)
	b32 := base32.StdEncoding.EncodeToString(raw)

	cases := map[string]string{
		"hex magnet":    "magnet:?xt=urn:btih:" + want + "&dn=x",
		"base32 magnet": "magnet:?xt=urn:btih:" + b32 + "&dn=x",
		"bare hex":      want,
		"bare base32":   b32,
	}
	for name, in := range cases {
		if got := Hash(in); got != want {
			t.Errorf("%s: Hash(%q) = %q, want %q", name, in, got, want)
		}
	}
	if got := Hash("not a magnet"); got != "" {
		t.Errorf("garbage: got %q, want empty", got)
	}
}
