package usenet

import (
	"context"
	"fmt"
	"io"
	"sort"
	"sync"

	"github.com/torrin-app/torrin/shared/usenet/decoder"
	"github.com/torrin-app/torrin/shared/usenet/nzb"
)

// ArticleFetcher is the narrow network boundary used by Reader. Implementations
// return the complete raw article body for a message ID and group.
type ArticleFetcher interface {
	Fetch(ctx context.Context, msgID, group string) ([]byte, error)
}

type segment struct {
	nzb.Segment
	start int64
	end   int64
}

// Reader exposes one NZB file as a lazy, seekable byte source. It retains only
// the most recently decoded segment, keeping memory bounded during long streams.
type Reader struct {
	ctx     context.Context
	fetcher ArticleFetcher
	group   string
	segs    []segment
	size    int64

	mu       sync.Mutex
	cacheIdx int
	cache    []byte
}

func NewReader(ctx context.Context, file nzb.File, fetcher ArticleFetcher) (*Reader, error) {
	if fetcher == nil {
		return nil, fmt.Errorf("usenet reader: nil article fetcher")
	}
	if len(file.Segments) == 0 {
		return nil, fmt.Errorf("usenet reader: file has no segments")
	}
	parts := append([]nzb.Segment(nil), file.Segments...)
	sort.SliceStable(parts, func(i, j int) bool { return parts[i].Number < parts[j].Number })
	r := &Reader{ctx: ctx, fetcher: fetcher, cacheIdx: -1}
	if len(file.Groups) > 0 {
		r.group = file.Groups[0]
	}
	for _, part := range parts {
		if part.Bytes <= 0 || part.MessageID == "" {
			return nil, fmt.Errorf("usenet reader: invalid segment %d", part.Number)
		}
		start := r.size
		r.size += part.Bytes
		if r.size < start {
			return nil, fmt.Errorf("usenet reader: file size overflow")
		}
		r.segs = append(r.segs, segment{Segment: part, start: start, end: r.size})
	}
	return r, nil
}

func (r *Reader) Size() int64 { return r.size }

func (r *Reader) ReadAt(p []byte, off int64) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if off < 0 {
		return 0, fmt.Errorf("usenet reader: negative offset")
	}
	if off >= r.size {
		return 0, io.EOF
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	written := 0
	idx := sort.Search(len(r.segs), func(i int) bool { return r.segs[i].end > off })
	for written < len(p) && idx < len(r.segs) {
		data, err := r.load(idx)
		if err != nil {
			return written, err
		}
		seg := r.segs[idx]
		at := off - seg.start
		n := copy(p[written:], data[at:])
		written += n
		off += int64(n)
		idx++
	}
	if written < len(p) {
		return written, io.EOF
	}
	return written, nil
}

func (r *Reader) load(idx int) ([]byte, error) {
	if r.cacheIdx == idx {
		return r.cache, nil
	}
	seg := r.segs[idx]
	raw, err := r.fetcher.Fetch(r.ctx, seg.MessageID, r.group)
	if err != nil {
		return nil, fmt.Errorf("fetch article %s: %w", seg.MessageID, err)
	}
	decoded, err := decoder.Decode(raw)
	if err != nil {
		return nil, fmt.Errorf("decode article %s: %w", seg.MessageID, err)
	}
	if int64(len(decoded.Data)) != seg.Bytes {
		return nil, fmt.Errorf("decode article %s: got %d bytes, want %d", seg.MessageID, len(decoded.Data), seg.Bytes)
	}
	if decoded.Begin > 0 && decoded.Begin-1 != seg.start {
		return nil, fmt.Errorf("decode article %s: begins at %d, want %d", seg.MessageID, decoded.Begin-1, seg.start)
	}
	r.cacheIdx, r.cache = idx, decoded.Data
	return r.cache, nil
}
