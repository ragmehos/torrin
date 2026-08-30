package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/torrin-app/torrin/shared/storage"
)

type encryptedSource interface {
	Head(ctx context.Context) (*storage.Object, error)
	Get(ctx context.Context, rng string) (*storage.Object, error)
}

type storageEncryptedSource struct {
	server *Server
	node   string
	key    string
}

func (s storageEncryptedSource) Head(ctx context.Context) (*storage.Object, error) {
	return s.server.store.HeadNode(ctx, s.node, s.key)
}

func (s storageEncryptedSource) Get(ctx context.Context, rng string) (*storage.Object, error) {
	return s.server.store.GetNode(ctx, s.node, s.key, rng)
}

func (s *Server) serveFileEnc(w http.ResponseWriter, r *http.Request, key string) {
	s.serveEncrypted(w, r, key, storageEncryptedSource{server: s, node: s.nodeFromReq(r), key: key})
}

func (s *Server) serveEncrypted(w http.ResponseWriter, r *http.Request, key string, source encryptedSource) {
	head, err := source.Head(r.Context())
	if err != nil {
		s.notFound(w, r, key, err)
		return
	}
	plainTotal, err := s.cipher.PlainSize(head.Size)
	if err != nil {
		slog.Warn("stream: bad encrypted object", "key", key, "size", head.Size, "err", err)
		httpError(w, 500, "bad object")
		return
	}

	h := w.Header()
	if head.ContentType != "" {
		h.Set("Content-Type", head.ContentType)
	}
	h.Set("Accept-Ranges", "bytes")
	s.setDownloadDisposition(w, r, key)
	s.recordView(r, key)
	if r.Method == http.MethodHead {
		h.Set("Content-Length", strconv.FormatInt(plainTotal, 10))
		h.Set("X-File-Size", strconv.FormatInt(plainTotal, 10))
		h.Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusOK)
		return
	}

	rangeHeader := r.Header.Get("Range")
	start, end, isRange := storage.ParseRange(rangeHeader, plainTotal)
	if rangeHeader != "" && !isRange {
		httpError(w, 416, "range not satisfiable")
		return
	}
	if !isRange {
		obj, err := source.Get(r.Context(), "")
		if err != nil {
			s.notFound(w, r, key, err)
			return
		}
		defer obj.Body.Close()
		h.Set("Content-Length", strconv.FormatInt(plainTotal, 10))
		w.WriteHeader(http.StatusOK)
		s.cipher.DecryptAll(w, obj.Body)
		return
	}

	plan, err := s.cipher.PlanRange(start, end+1, plainTotal)
	if err != nil {
		httpError(w, 416, "range not satisfiable")
		return
	}
	encRange := fmt.Sprintf("bytes=%d-%d", plan.EncStart, plan.EncEnd-1)
	obj, err := source.Get(r.Context(), encRange)
	if err != nil {
		s.notFound(w, r, key, err)
		return
	}
	defer obj.Body.Close()

	h.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, plainTotal))
	h.Set("Content-Length", strconv.FormatInt(end-start+1, 10))
	w.WriteHeader(http.StatusPartialContent)
	s.cipher.DecryptRange(w, obj.Body, plan)
}
