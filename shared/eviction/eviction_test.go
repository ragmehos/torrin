package eviction

import (
	"context"
	"testing"

	"github.com/torrin-app/torrin/shared/jobs"
)

type fakeRepo struct {
	candidates []jobs.EvictionCandidate
	total      int64
	evicted    map[string]bool
}

func (f *fakeRepo) GetEvictionCandidates(context.Context) ([]jobs.EvictionCandidate, error) {
	return f.candidates, nil
}
func (f *fakeRepo) GetTotalCachedSize(context.Context) (int64, error) { return f.total, nil }
func (f *fakeRepo) ListByInfoHash(_ context.Context, h string) ([]*jobs.Job, error) {
	return []*jobs.Job{{ID: h, InfoHash: h}}, nil
}
func (f *fakeRepo) Update(_ context.Context, j *jobs.Job) error {
	if j.Status == jobs.StatusEvicted {
		f.evicted[j.InfoHash] = true
	}
	return nil
}

type fakeStorage struct{ deleted map[string]bool }

func (f *fakeStorage) DeletePrefix(_ context.Context, prefix string) error {
	f.deleted[prefix] = true
	return nil
}

func TestTTLEviction(t *testing.T) {
	repo := &fakeRepo{
		evicted: map[string]bool{},
		candidates: []jobs.EvictionCandidate{
			{InfoHash: "never_old", AccessCount: 0, DaysSinceAccess: 10},     // > 7 → evict
			{InfoHash: "never_fresh", AccessCount: 0, DaysSinceAccess: 5},    // < 7 → keep
			{InfoHash: "popular", AccessCount: 15, DaysSinceAccess: 10},      // < 45 → keep
			{InfoHash: "stale_mid", AccessCount: 3, DaysSinceAccess: 20},     // > 14 → evict
			{InfoHash: "big", FileSize: 60_000_000_000, DaysSinceAccess: 50}, // large, > 45 → evict
		},
	}
	store := &fakeStorage{deleted: map[string]bool{}}
	New(repo, store, DefaultPolicy).RunDaily(context.Background())

	for _, h := range []string{"never_old", "stale_mid", "big"} {
		if !store.deleted[h+"/"] || !repo.evicted[h] {
			t.Errorf("%s should be evicted", h)
		}
	}
	for _, h := range []string{"never_fresh", "popular"} {
		if store.deleted[h+"/"] {
			t.Errorf("%s should NOT be evicted", h)
		}
	}
}

func TestBudgetPassRespectsGrace(t *testing.T) {
	repo := &fakeRepo{
		evicted: map[string]bool{},
		total:   400_000_000_000, // over the 300GB cap
		candidates: []jobs.EvictionCandidate{
			{InfoHash: "fresh_big", FileSize: 200_000_000_000, AccessCount: 20, DaysSinceAccess: 1}, // within grace → skip
		},
	}
	store := &fakeStorage{deleted: map[string]bool{}}
	New(repo, store, DefaultPolicy).RunDaily(context.Background())

	if store.deleted["fresh_big/"] {
		t.Error("content inside grace window must not be budget-evicted")
	}
}
