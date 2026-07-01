package providers

import (
	"context"
	"log/slog"
)

func FirstAvailable(ctx context.Context, provs []Provider, magnet, infoHash string) (*Result, Provider, error) {
	var firstErr error
	for _, p := range provs {
		res, err := p.Fetch(ctx, magnet, infoHash)
		if err != nil {
			slog.Info("debrid miss", "provider", p.Name(), "reason", err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if res != nil {
			return res, p, nil
		}
		slog.Info("debrid miss", "provider", p.Name(), "reason", "not cached")
	}
	return nil, nil, firstErr
}
