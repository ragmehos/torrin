package jobs

import (
	"context"
	"time"
)

type UserStats struct {
	TotalDownloads int   `json:"total_downloads"`
	ActiveJobs     int   `json:"active_jobs"`
	CompletedJobs  int   `json:"completed_jobs"`
	FailedJobs     int   `json:"failed_jobs"`
	TotalBytes     int64 `json:"total_bytes"`
	TotalAccesses  int64 `json:"total_accesses"`
}

type HistoryEntry struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	InfoHash     string `json:"info_hash"`
	Status       Status `json:"status"`
	FileSize     int64  `json:"file_size"`
	AccessCount  int64  `json:"access_count"`
	CreatedAt    string `json:"created_at"`
	LastAccessed string `json:"last_accessed,omitempty"`
}

func (p *Postgres) GetUserStats(ctx context.Context, userID string) (*UserStats, error) {
	st := &UserStats{}
	p.pool.QueryRow(ctx, `SELECT COUNT(*) FROM jobs WHERE user_id=$1`, userID).Scan(&st.TotalDownloads)
	p.pool.QueryRow(ctx, `SELECT COUNT(*) FROM jobs WHERE user_id=$1 AND status IN `+activeStates, userID).Scan(&st.ActiveJobs)
	p.pool.QueryRow(ctx, `SELECT COUNT(*) FROM jobs WHERE user_id=$1 AND status='complete'`, userID).Scan(&st.CompletedJobs)
	p.pool.QueryRow(ctx, `SELECT COUNT(*) FROM jobs WHERE user_id=$1 AND status='failed'`, userID).Scan(&st.FailedJobs)
	p.pool.QueryRow(ctx, `SELECT COALESCE(SUM(file_size),0) FROM jobs WHERE user_id=$1 AND status='complete'`, userID).Scan(&st.TotalBytes)
	p.pool.QueryRow(ctx, `SELECT COALESCE(SUM(access_count),0) FROM jobs WHERE user_id=$1`, userID).Scan(&st.TotalAccesses)
	return st, nil
}

func (p *Postgres) GetUserHistory(ctx context.Context, userID string, limit int) ([]HistoryEntry, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := p.pool.Query(ctx, `
		SELECT id, name, info_hash, status, file_size, access_count, created_at, last_accessed_at
		FROM jobs WHERE user_id=$1 AND status IN ('complete','evicted')
		ORDER BY created_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []HistoryEntry{}
	for rows.Next() {
		var h HistoryEntry
		var status string
		var created time.Time
		var last *time.Time
		if err := rows.Scan(&h.ID, &h.Name, &h.InfoHash, &status, &h.FileSize, &h.AccessCount, &created, &last); err != nil {
			return nil, err
		}
		h.Status = Status(status)
		h.CreatedAt = created.UTC().Format(time.RFC3339)
		if last != nil {
			h.LastAccessed = last.UTC().Format(time.RFC3339)
		}
		out = append(out, h)
	}
	return out, rows.Err()
}
