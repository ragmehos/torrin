package providers

import (
	"context"
	"testing"
)

var (
	_ Provider = (*realdebrid)(nil)
	_ Provider = (*alldebrid)(nil)
	_ Provider = (*premiumize)(nil)
	_ Provider = (*torbox)(nil)
)

type fakeProvider struct {
	name string
	res  *Result
	err  error
}

func (f *fakeProvider) Name() string                                           { return f.name }
func (f *fakeProvider) Fetch(context.Context, string, string) (*Result, error) { return f.res, f.err }
func (f *fakeProvider) Release(context.Context, string) error                  { return nil }

func TestFirstAvailableOrder(t *testing.T) {
	hit := &Result{Name: "movie"}
	provs := []Provider{
		&fakeProvider{name: "a"},
		&fakeProvider{name: "b", res: hit},
		&fakeProvider{name: "c", res: &Result{Name: "other"}},
	}
	res, p, err := FirstAvailable(context.Background(), provs, "magnet", "hash")
	if err != nil {
		t.Fatal(err)
	}
	if res != hit {
		t.Fatalf("got %v, want first hit", res)
	}
	if p.Name() != "b" {
		t.Fatalf("got provider %s, want b", p.Name())
	}
}

func TestFirstAvailableMiss(t *testing.T) {
	provs := []Provider{&fakeProvider{name: "a"}, &fakeProvider{name: "b"}}
	res, _, err := FirstAvailable(context.Background(), provs, "m", "h")
	if err != nil || res != nil {
		t.Fatalf("want clean miss, got res=%v err=%v", res, err)
	}
}

func TestIsVideoFile(t *testing.T) {
	if !isVideoFile("Movie.2020.1080p.mkv") {
		t.Error("mkv should be video")
	}
	if isVideoFile("readme.txt") {
		t.Error("txt should not be video")
	}
	if !isVideoFile("clip.MP4") {
		t.Error("uppercase ext should match")
	}
}

func TestTotalFromContentRange(t *testing.T) {
	if got := totalFromContentRange("bytes 0-99/12345"); got != 12345 {
		t.Errorf("got %d, want 12345", got)
	}
	if got := totalFromContentRange("bogus"); got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}
