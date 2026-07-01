package server

import "sync"

const maxZipsPerUser = 2

type zipGuard struct {
	mu       sync.Mutex
	inflight map[string]int
	max      int
}

func newZipGuard(max int) *zipGuard {
	return &zipGuard{inflight: map[string]int{}, max: max}
}

func (g *zipGuard) acquire(user string) bool {
	if user == "" || g.max <= 0 {
		return true
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.inflight[user] >= g.max {
		return false
	}
	g.inflight[user]++
	return true
}

func (g *zipGuard) release(user string) {
	if user == "" || g.max <= 0 {
		return
	}
	g.mu.Lock()
	g.inflight[user]--
	if g.inflight[user] <= 0 {
		delete(g.inflight, user)
	}
	g.mu.Unlock()
}
