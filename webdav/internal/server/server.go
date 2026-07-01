package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/torrin-app/torrin/shared/auth"
	"github.com/torrin-app/torrin/shared/jobs"
	"github.com/torrin-app/torrin/shared/manifest"
	"github.com/torrin-app/torrin/shared/storage"
	"github.com/torrin-app/torrin/shared/video"
)

type Server struct {
	users *auth.Store
	jobs  jobs.Repository
	store *storage.Client
}

func New(users *auth.Store, j jobs.Repository, store *storage.Client) *Server {
	return &Server{users: users, jobs: j, store: store}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.serve)
	return mux
}

func (s *Server) serve(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("DAV", "1, 2")
	w.Header().Set("Allow", "OPTIONS, GET, HEAD, PROPFIND")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	user, err := s.authenticate(r)
	if err != nil {
		w.Header().Set("WWW-Authenticate", `Basic realm="Torrin"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	path := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/"), "/")
	switch r.Method {
	case "PROPFIND":
		s.propfind(w, r, user.ID, path)
	case http.MethodGet, http.MethodHead:
		s.get(w, r, user.ID, path)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) authenticate(r *http.Request) (*auth.User, error) {
	u, p, ok := r.BasicAuth()
	if !ok {
		return nil, fmt.Errorf("no credentials")
	}
	apiKey := u
	if strings.HasPrefix(p, "tr_") {
		apiKey = p
	}
	if !strings.HasPrefix(apiKey, "tr_") {
		return nil, fmt.Errorf("invalid api key")
	}
	user, err := s.users.GetByAPIKey(r.Context(), apiKey)
	if err != nil {
		return nil, fmt.Errorf("invalid api key")
	}
	if time.Now().After(user.ExpiresAt) {
		return nil, fmt.Errorf("subscription expired")
	}
	return user, nil
}

type fileRef struct {
	name string
	key  string
	hash string
	size int64
	mod  time.Time
}

func (s *Server) files(ctx context.Context, userID string) []fileRef {
	list, _ := s.jobs.ListByUser(ctx, userID, 200)
	var out []fileRef
	for _, j := range list {
		if j.Status != jobs.StatusComplete {
			continue
		}
		for i, f := range j.Files {
			out = append(out, fileRef{
				name: f.Name, key: manifest.Key(j.InfoHash, i, f.Name),
				hash: j.InfoHash, size: f.Size, mod: j.UpdatedAt,
			})
		}
	}
	return out
}

func (s *Server) propfind(w http.ResponseWriter, r *http.Request, userID, path string) {
	files := s.files(r.Context(), userID)
	if path == "" {
		entries := []propEntry{{Href: "/", IsCollection: true, Name: "Torrin"}}
		for _, f := range files {
			entries = append(entries, propEntry{Href: "/" + f.name, Name: f.name, Size: f.size, Modified: f.mod})
		}
		writePropfind(w, entries)
		return
	}
	for _, f := range files {
		if f.name == path {
			writePropfind(w, []propEntry{{Href: "/" + f.name, Name: f.name, Size: f.size, Modified: f.mod}})
			return
		}
	}
	http.Error(w, "Not Found", http.StatusNotFound)
}

func (s *Server) get(w http.ResponseWriter, r *http.Request, userID, path string) {
	if path == "" {
		http.Error(w, "cannot download a directory", http.StatusBadRequest)
		return
	}
	for _, f := range s.files(r.Context(), userID) {
		if f.name == path {
			s.jobs.RecordView(r.Context(), f.hash, userID)
			http.Redirect(w, r, s.store.SignURL(f.key, 4*time.Hour), http.StatusTemporaryRedirect)
			return
		}
	}
	http.Error(w, "Not Found", http.StatusNotFound)
}

type propEntry struct {
	Href         string
	IsCollection bool
	Name         string
	Size         int64
	Modified     time.Time
}

func writePropfind(w http.ResponseWriter, entries []propEntry) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusMultiStatus)

	io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?>`)
	io.WriteString(w, `<D:multistatus xmlns:D="DAV:">`)
	for _, e := range entries {
		io.WriteString(w, `<D:response><D:href>`+xmlEscape(e.Href)+`</D:href>`)
		io.WriteString(w, `<D:propstat><D:prop>`)
		io.WriteString(w, `<D:displayname>`+xmlEscape(e.Name)+`</D:displayname>`)
		if e.IsCollection {
			io.WriteString(w, `<D:resourcetype><D:collection/></D:resourcetype>`)
		} else {
			io.WriteString(w, `<D:resourcetype/>`)
			io.WriteString(w, fmt.Sprintf(`<D:getcontentlength>%d</D:getcontentlength>`, e.Size))
			io.WriteString(w, `<D:getcontenttype>`+video.ContentType(e.Name)+`</D:getcontenttype>`)
			if !e.Modified.IsZero() {
				io.WriteString(w, `<D:getlastmodified>`+e.Modified.UTC().Format(http.TimeFormat)+`</D:getlastmodified>`)
			}
		}
		io.WriteString(w, `</D:prop><D:status>HTTP/1.1 200 OK</D:status></D:propstat></D:response>`)
	}
	io.WriteString(w, `</D:multistatus>`)
}

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}
