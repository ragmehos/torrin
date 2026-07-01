package screen

import (
	"context"

	"github.com/torrin-app/torrin/shared/jobs"
	"github.com/torrin-app/torrin/shared/safety"
)

type BanFunc func(ctx context.Context, userID, reason string)

func Blocked(ctx context.Context, ban BanFunc, job *jobs.Job, names ...string) bool {
	v := safety.Screen(append([]string{job.Name, job.Magnet}, names...)...)
	if !v.Blocked {
		return false
	}
	if v.Ban && ban != nil && job.UserID != "" && job.UserID != "system" {
		ban(ctx, job.UserID, v.Reason)
	}
	return true
}
