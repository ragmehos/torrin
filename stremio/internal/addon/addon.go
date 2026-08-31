package addon

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/torrin-app/torrin/shared/auth"
	"github.com/torrin-app/torrin/shared/cinemeta"
	"github.com/torrin-app/torrin/shared/georoute"
	"github.com/torrin-app/torrin/shared/jobs"
	"github.com/torrin-app/torrin/shared/manifest"
	"github.com/torrin-app/torrin/shared/storage"
	"github.com/torrin-app/torrin/shared/stremioid"
)

type Server struct {
	users *auth.Store
	jobs  *jobs.Postgres
	store *storage.Client
	meta  *cinemeta.Client
}

func New(users *auth.Store, j *jobs.Postgres, store *storage.Client) *Server {
	return &Server{users: users, jobs: j, store: store, meta: cinemeta.NewClient()}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{apiKey}/manifest.json", s.manifest)
	mux.HandleFunc("GET /{apiKey}/stream/{type}/{id}", s.stream)
	return cors(mux)
}

func (s *Server) manifest(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, map[string]any{
		"id":          "app.torrin.stremio",
		"version":     "1.0.0",
		"name":        "Torrin",
		"description": "Stream your cached media via Torrin",
		"types":       []string{"movie", "series"},
		"catalogs":    []any{},
		"resources":   []string{"stream"},
		"idPrefixes":  []string{"tt"},
		"behaviorHints": map[string]any{
			"configurable":          false,
			"configurationRequired": false,
		},
	})
}

func (s *Server) stream(w http.ResponseWriter, r *http.Request) {
	empty := map[string]any{"streams": []any{}}
	apiKey := r.PathValue("apiKey")
	contentID := strings.TrimSuffix(r.PathValue("id"), ".json")
	if apiKey == "" || contentID == "" {
		writeJSON(w, 200, empty)
		return
	}

	user, err := s.users.GetByAPIKey(r.Context(), apiKey)
	if err != nil || user == nil || user.Banned || user.IsPaused() || time.Now().After(user.ExpiresAt) {
		writeJSON(w, 200, empty)
		return
	}

	byos := s.userHasBYOS(r.Context(), user.ID)
	id := stremioid.Parse(contentID)
	var streams []map[string]any
	if id.InfoHash != "" {
		streams = append(streams, s.byHash(r, id.InfoHash, user.ID, byos)...)
	}
	if id.IMDBID != "" {
		streams = append(streams, s.byLibrary(r, r.PathValue("type"), id, user.ID, byos)...)
	}

	if len(streams) == 0 {
		writeJSON(w, 200, empty)
		return
	}
	if id.InfoHash != "" {
		s.jobs.RecordView(r.Context(), id.InfoHash, user.ID)
	}
	slog.Info("stremio: served", "user", user.ID, "id", contentID, "streams", len(streams))
	writeJSON(w, 200, map[string]any{"streams": streams})
}

func (s *Server) byHash(r *http.Request, infoHash, userID string, byos bool) []map[string]any {
	data, err := s.store.GetBytes(r.Context(), manifest.Path(infoHash))
	if err != nil {
		return nil
	}
	man, err := manifest.Parse(data)
	if err != nil {
		slog.Warn("stremio: bad manifest", "hash", infoHash, "err", err)
		return nil
	}
	var out []map[string]any
	for _, f := range man.Files {
		out = append(out, entry(f.FileName, s.streamURL(r, infoHash, f.DirectURL, userID, byos, f.Enc)))
	}
	return out
}

func (s *Server) userHasBYOS(ctx context.Context, userID string) bool {
	creds, err := s.users.GetStorageCreds(ctx, userID)
	return err == nil && creds != nil && creds.Enabled && creds.IsRclone()
}

func (s *Server) streamURL(r *http.Request, infoHash, key, userID string, byos, enc bool) string {
	var u string
	if byos {
		u = s.store.SignURLNodeUser("", key, userID, 24*time.Hour) + "&byos=1"
	} else {
		u = s.store.SignURLNode(s.jobs.NodeForInfoHash(r.Context(), infoHash), key, 24*time.Hour)
	}
	u += manifest.StreamQuery(infoHash, enc)
	return georoute.URL(r, u)
}

func (s *Server) byLibrary(r *http.Request, contentType string, id stremioid.ID, userID string, byos bool) []map[string]any {
	ctx := r.Context()
	seen := map[string]bool{}
	var out []map[string]any
	add := func(list []*jobs.Job) {
		for _, j := range list {
			if seen[j.InfoHash] {
				continue
			}
			files := libraryFiles(j, id)
			if len(files) == 0 {
				continue
			}
			seen[j.InfoHash] = true
			for _, f := range files {
				key := manifest.ResolveKey(j.InfoHash, f.Index, f.Key, f.Name)
				out = append(out, entry(f.Name, s.streamURL(r, j.InfoHash, key, userID, byos, f.Enc)))
			}
		}
	}

	byImdb, _ := s.jobs.ListByIMDB(ctx, id.IMDBID)
	add(byImdb)

	if byos {
		byosOwn, _ := s.jobs.ListUserByosByIMDB(ctx, userID, id.IMDBID)
		add(byosOwn)
	}

	if title, err := s.meta.Title(ctx, id.IMDBID, contentType); err == nil {
		if norm := jobs.NormTitle(title); norm != "" {
			byTitle, _ := s.jobs.ListByTitleNorm(ctx, norm)
			add(byTitle)
		}
	}
	return out
}

func libraryFiles(j *jobs.Job, id stremioid.ID) []jobs.File {
	return jobs.FilesForEpisode(j, j.Files, id.Season, id.Episode)
}

func entry(title, url string) map[string]any {
	return map[string]any{
		"name":  "Torrin",
		"title": title,
		"url":   url,
		"behaviorHints": map[string]any{
			"notWebReady": strings.HasSuffix(title, ".mkv"),
		},
	}
}

func cors(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
