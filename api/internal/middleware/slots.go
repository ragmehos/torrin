package middleware

import (
	"context"
	"sync"

	"github.com/torrin-app/torrin/shared/jobs"
	"github.com/torrin-app/torrin/shared/plans"
)

type SlotTracker struct {
	repo  jobs.Repository
	locks sync.Map
}

func NewSlotTracker(repo jobs.Repository) *SlotTracker {
	return &SlotTracker{repo: repo}
}

func (st *SlotTracker) userLock(userID string) *sync.Mutex {
	v, _ := st.locks.LoadOrStore(userID, &sync.Mutex{})
	return v.(*sync.Mutex)
}

func (st *SlotTracker) ActiveSlots(ctx context.Context, userID string) int {
	n, _ := st.repo.ActiveCount(ctx, userID)
	return n
}

func (st *SlotTracker) Acquire(ctx context.Context, userID string, plan plans.Plan) bool {
	mu := st.userLock(userID)
	mu.Lock()
	if st.ActiveSlots(ctx, userID) < plan.MaxConcurrent {
		return true
	}
	mu.Unlock()
	return false
}

func (st *SlotTracker) Release(userID string) {
	st.userLock(userID).Unlock()
}
