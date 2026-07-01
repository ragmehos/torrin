package scrape

import (
	"context"
	"testing"
	"time"
)

func TestCheckServesCache(t *testing.T) {
	// All hashes are fresh in cache → Check must NOT scrape (no network).
	c := &Client{cache: map[string]entry{
		"h1": {res: Result{Seeders: 10, Leechers: 2}, exp: time.Now().Add(time.Hour)},
		"h2": {res: Result{Seeders: 5}, exp: time.Now().Add(time.Hour)},
	}}
	out := c.Check(context.Background(), []string{"h1", "h2"})
	if out["h1"].Seeders != 10 || out["h1"].Leechers != 2 || out["h2"].Seeders != 5 {
		t.Errorf("cache not served: %+v", out)
	}
}

func TestBatchScrapeEmpty(t *testing.T) {
	if BatchScrape(context.Background(), nil) != nil {
		t.Error("empty hashes should return nil")
	}
}
