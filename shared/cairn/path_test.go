package cairn

import (
	"strings"
	"testing"
)

func TestStreamPathRoundTrip(t *testing.T) {
	hash := strings.Repeat("a", 40)
	path := StreamPath(hash, 2, "folder/Movie Name.mkv")
	if path != hash+"/cairn/2/Movie Name.mkv" {
		t.Fatalf("path = %q", path)
	}
	gotHash, idx, name, ok := ParseStreamPath(path)
	if !ok || gotHash != hash || idx != 2 || name != "Movie Name.mkv" {
		t.Fatalf("parse = %q %d %q %v", gotHash, idx, name, ok)
	}
}

func TestParseStreamPathRejectsMalformed(t *testing.T) {
	for _, path := range []string{"", "abc/cairn/0/x", strings.Repeat("g", 40) + "/cairn/0/x", strings.Repeat("a", 40) + "/cairn/-1/x"} {
		if _, _, _, ok := ParseStreamPath(path); ok {
			t.Fatalf("accepted %q", path)
		}
	}
}
