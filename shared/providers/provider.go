package providers

import (
	"context"

	"github.com/torrin-app/torrin/shared/video"
)

type Provider interface {
	Name() string
	Fetch(ctx context.Context, magnet, infoHash string) (*Result, error)
	Release(ctx context.Context, handle string) error
}

type Result struct {
	Name   string
	Handle string
	Files  []Link
}

type Link struct {
	Name  string
	Size  int64
	URL   string
	Renew func(context.Context) (string, error)
}

func isVideoFile(name string) bool {
	return video.IsVideo(name)
}
