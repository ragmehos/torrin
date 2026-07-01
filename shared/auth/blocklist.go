package auth

import (
	"context"
	"time"
)

func (s *Store) BanUser(ctx context.Context, userID, reason string) error {
	_, err := s.pool.Exec(ctx, `UPDATE users SET banned=TRUE, ban_reason=$1, updated_at=$2 WHERE id=$3`,
		reason, time.Now(), userID)
	return err
}

func (s *Store) GetBlocklist(ctx context.Context) (hard, soft []string, err error) {
	rows, err := s.pool.Query(ctx, `SELECT term, tier FROM blocklist_terms`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var term, tier string
		if rows.Scan(&term, &tier) != nil {
			continue
		}
		if tier == "soft" {
			soft = append(soft, term)
		} else {
			hard = append(hard, term)
		}
	}
	return hard, soft, rows.Err()
}
