package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"github.com/torrin-app/torrin/shared/cairn"
	"github.com/torrin-app/torrin/shared/storage"
	sharedusenet "github.com/torrin-app/torrin/shared/usenet"
	"github.com/torrin-app/torrin/shared/usenet/nzb"
	"github.com/torrin-app/torrin/shared/video"
)

func (s *Server) serveCairn(w http.ResponseWriter, r *http.Request, key string) {
	hash, index, requestedName, ok := cairn.ParseStreamPath(key)
	if !ok {
		httpError(w, 404, "not found")
		return
	}
	if s.cairnStore == nil || s.cairnFetch == nil {
		httpError(w, 503, "cairn streaming unavailable")
		return
	}
	if r.Method == http.MethodGet {
		identity := cairnStreamIdentity(r)
		switch s.cairnSlots.acquire(identity) {
		case streamLimitUser:
			w.Header().Set("Retry-After", "5")
			httpError(w, http.StatusTooManyRequests, "cairn user stream limit reached")
			return
		case streamLimitTotal:
			w.Header().Set("Retry-After", "5")
			httpError(w, http.StatusServiceUnavailable, "cairn stream capacity reached")
			return
		default:
			defer s.cairnSlots.release(identity)
		}
	}
	data, err := s.cairnStore.GetBytes(r.Context(), nzb.StorageKey(hash))
	if err != nil {
		s.notFound(w, r, key, err)
		return
	}
	parsed, err := nzb.ParseBytes(data)
	if err != nil {
		s.notFound(w, r, key, fmt.Errorf("invalid cairn nzb: %w", err))
		return
	}
	if index >= len(parsed.Files) {
		s.notFound(w, r, key, fmt.Errorf("invalid cairn file index %d", index))
		return
	}
	file := parsed.Files[index]
	name := file.Filename
	if name == "" {
		name = file.Subject
	}
	name = filepath.Base(name)
	if name != requestedName {
		httpError(w, 404, "not found")
		return
	}
	reader, err := sharedusenet.NewReader(r.Context(), file, s.cairnFetch)
	if err != nil {
		s.notFound(w, r, key, err)
		return
	}

	h := w.Header()
	h.Set("Content-Type", video.ContentType(name))
	h.Set("Cache-Control", "no-store")
	if s.cipher != nil && r.URL.Query().Get("enc") == "1" {
		s.serveEncrypted(w, r, key, readerEncryptedSource{reader: reader})
		return
	}
	s.setDownloadDisposition(w, r, key)
	s.recordView(r, key)
	http.ServeContent(w, r, name, time.Time{}, io.NewSectionReader(reader, 0, reader.Size()))
}

type readerEncryptedSource struct{ reader *sharedusenet.Reader }

func (s readerEncryptedSource) Head(context.Context) (*storage.Object, error) {
	return &storage.Object{Size: s.reader.Size()}, nil
}

func (s readerEncryptedSource) Get(_ context.Context, rng string) (*storage.Object, error) {
	start, end := int64(0), s.reader.Size()-1
	if rng != "" {
		var ok bool
		start, end, ok = storage.ParseRange(rng, s.reader.Size())
		if !ok {
			return nil, fmt.Errorf("invalid encrypted range %q", rng)
		}
	}
	length := end - start + 1
	return &storage.Object{
		Body:         io.NopCloser(io.NewSectionReader(s.reader, start, length)),
		Size:         length,
		ContentRange: fmt.Sprintf("bytes %d-%d/%d", start, end, s.reader.Size()),
	}, nil
}
