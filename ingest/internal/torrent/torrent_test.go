package torrent

import (
	"context"
	"testing"

	"github.com/torrin-app/torrin/shared/jobs"
)

type fakeRepo struct {
	jobs.Repository
	byStatus map[jobs.Status][]*jobs.Job
}

func (f *fakeRepo) ListByStatus(_ context.Context, s jobs.Status) ([]*jobs.Job, error) {
	return f.byStatus[s], nil
}

func TestActiveHashesProtectsLiveJobs(t *testing.T) {
	v2 := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	r := &Runner{repo: &fakeRepo{byStatus: map[jobs.Status][]*jobs.Job{
		jobs.StatusDownloading: {{InfoHash: "AABBCC"}},
		jobs.StatusPublishing: {{InfoHash: "0000000000000000000000000000000000000000",
			Magnet: "magnet:?xt=urn:btih:0000000000000000000000000000000000000000&xt=urn:btmh:1220" + v2}},
		jobs.StatusComplete: {{InfoHash: "deadbeef"}},
	}}}
	keep := r.activeHashes(context.Background())
	if !keep["aabbcc"] {
		t.Error("downloading job not protected")
	}
	if !keep[v2[:40]] {
		t.Error("v2 hash of active job not protected")
	}
	if keep["deadbeef"] {
		t.Error("completed job should not be protected (reapable)")
	}
}

func TestExtractV2Hash(t *testing.T) {
	v2 := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" // 64 hex (sha256)
	magnet := "magnet:?xt=urn:btih:0000000000000000000000000000000000000000&xt=urn:btmh:1220" + v2
	if got := extractV2Hash(magnet); got != v2[:40] {
		t.Errorf("got %q, want %q", got, v2[:40])
	}
	if got := extractV2Hash("magnet:?xt=urn:btih:abc"); got != "" {
		t.Errorf("no btmh should yield empty, got %q", got)
	}
}
