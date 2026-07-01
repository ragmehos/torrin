package scrape

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/etix/goscrape"
)

const (
	cacheTTL    = 24 * time.Hour
	maxHashes   = 74
	trackerWait = 5 * time.Second
)

var defaultTrackers = []string{
	"udp://tracker.opentrackr.org:1337/announce",
	"udp://open.stealth.si:80/announce",
	"udp://tracker.torrent.eu.org:451/announce",
	"udp://tracker.opentorrent.top:6969/announce",
	"udp://open.demonii.com:1337/announce",
	"udp://explodie.org:6969/announce",
	"udp://tracker.dler.com:6969/announce",
	"udp://tracker-udp.gbitt.info:80/announce",
}

type Result struct {
	Seeders   int `json:"seeders"`
	Leechers  int `json:"leechers"`
	Completed int `json:"completed"`
}

type entry struct {
	res Result
	exp time.Time
}

type Client struct {
	mu    sync.Mutex
	cache map[string]entry
}

func New() *Client {
	c := &Client{cache: make(map[string]entry)}
	go c.sweep()
	return c
}

func (c *Client) sweep() {
	t := time.NewTicker(time.Hour)
	defer t.Stop()
	for range t.C {
		now := time.Now()
		c.mu.Lock()
		for k, e := range c.cache {
			if now.After(e.exp) {
				delete(c.cache, k)
			}
		}
		c.mu.Unlock()
	}
}

func (c *Client) Check(ctx context.Context, hashes []string) map[string]Result {
	if len(hashes) > maxHashes {
		hashes = hashes[:maxHashes]
	}
	out := make(map[string]Result)
	var miss []string
	now := time.Now()

	c.mu.Lock()
	for _, h := range hashes {
		if e, ok := c.cache[h]; ok && now.Before(e.exp) {
			out[h] = e.res
		} else {
			miss = append(miss, h)
		}
	}
	c.mu.Unlock()

	if len(miss) == 0 {
		return out
	}

	fresh := BatchScrape(ctx, miss)
	exp := now.Add(cacheTTL)
	c.mu.Lock()
	for h, r := range fresh {
		c.cache[h] = entry{res: r, exp: exp}
		out[h] = r
	}
	c.mu.Unlock()
	return out
}

func BatchScrape(ctx context.Context, hashes []string) map[string]Result {
	if len(hashes) == 0 {
		return nil
	}
	if len(hashes) > maxHashes {
		hashes = hashes[:maxHashes]
	}
	infohashes := make([][]byte, len(hashes))
	for i, h := range hashes {
		infohashes[i] = []byte(h)
	}

	results := make(map[string]Result)
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, tracker := range defaultTrackers {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			s, err := goscrape.New(addr)
			if err != nil {
				return
			}
			done := make(chan struct{})
			var res []*goscrape.ScrapeResult
			var scrapeErr error
			go func() {
				res, scrapeErr = s.Scrape(infohashes...)
				close(done)
			}()
			select {
			case <-done:
			case <-time.After(trackerWait):
				slog.Debug("scrape: tracker timeout", "tracker", addr)
				return
			case <-ctx.Done():
				return
			}
			if scrapeErr != nil {
				return
			}
			mu.Lock()
			for _, r := range res {
				hash := string(r.Infohash)
				if e, ok := results[hash]; !ok || int(r.Seeders) > e.Seeders {
					results[hash] = Result{Seeders: int(r.Seeders), Leechers: int(r.Leechers), Completed: int(r.Completed)}
				}
			}
			mu.Unlock()
		}(tracker)
	}
	wg.Wait()
	return results
}
