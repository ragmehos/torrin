package nzb

import "testing"

func TestTotalSizeAndName(t *testing.T) {
	n := &NZB{
		Meta:  map[string]string{"name": "My.Release"},
		Files: []File{{Segments: []Segment{{Bytes: 100}, {Bytes: 50}}}},
	}
	if n.TotalSize() != 150 {
		t.Errorf("total = %d, want 150", n.TotalSize())
	}
	if n.Name() != "My.Release" {
		t.Errorf("name = %q", n.Name())
	}
}

func TestHashStable(t *testing.T) {
	a := &NZB{Files: []File{{Segments: []Segment{{MessageID: "b"}, {MessageID: "a"}}}}}
	b := &NZB{Files: []File{{Segments: []Segment{{MessageID: "a"}, {MessageID: "b"}}}}}
	if Hash(a) != Hash(b) {
		t.Error("hash should be order-independent")
	}
	if len(Hash(a)) != 40 {
		t.Errorf("hash length = %d, want 40", len(Hash(a)))
	}
}
