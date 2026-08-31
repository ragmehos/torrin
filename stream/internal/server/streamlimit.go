package server

import (
	"net/http"
	"sync"
)

type streamLimit uint8

const (
	streamAllowed streamLimit = iota
	streamLimitUser
	streamLimitTotal
)

// streamGuard bounds long-lived direct Cairn responses. The global ceiling
// protects the NNTP pool while the per-user ceiling preserves fair access.
type streamGuard struct {
	mu         sync.Mutex
	inflight   map[string]int
	total      int
	maxTotal   int
	maxPerUser int
}

func newStreamGuard(total, perUser int) *streamGuard {
	if total < 1 {
		total = 1
	}
	if perUser < 1 {
		perUser = 1
	}
	if perUser > total {
		perUser = total
	}
	return &streamGuard{inflight: map[string]int{}, maxTotal: total, maxPerUser: perUser}
}

func (g *streamGuard) acquire(identity string) streamLimit {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.inflight[identity] >= g.maxPerUser {
		return streamLimitUser
	}
	if g.total >= g.maxTotal {
		return streamLimitTotal
	}
	g.inflight[identity]++
	g.total++
	return streamAllowed
}

func (g *streamGuard) release(identity string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.inflight[identity] == 0 {
		return
	}
	g.inflight[identity]--
	g.total--
	if g.inflight[identity] == 0 {
		delete(g.inflight, identity)
	}
}

// New links carry a user ID covered by the URL signature. The signature is a
// stable compatibility key for already-issued links that predate user binding.
func cairnStreamIdentity(r *http.Request) string {
	if user := r.URL.Query().Get("u"); user != "" {
		return "user:" + user
	}
	return "legacy:" + r.URL.Query().Get("sig")
}
