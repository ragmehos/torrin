package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/torrin-app/torrin/api/internal/middleware"
	"github.com/torrin-app/torrin/api/internal/web"
	"github.com/torrin-app/torrin/shared/auth"
	"github.com/torrin-app/torrin/shared/plans"
	"github.com/torrin-app/torrin/shared/usenet/indexer"
)

func (s *Server) registerUsenetSearchRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
	mux.Handle("POST /api/usenet/indexer", authMW(http.HandlerFunc(s.setIndexer)))
	mux.Handle("GET /api/usenet/indexer", authMW(http.HandlerFunc(s.getIndexer)))
	mux.Handle("DELETE /api/usenet/indexer", authMW(http.HandlerFunc(s.delIndexer)))
	mux.Handle("POST /api/usenet/indexer/test", authMW(http.HandlerFunc(s.testIndexer)))
	mux.Handle("GET /api/usenet/search", authMW(http.HandlerFunc(s.usenetSearch)))
	mux.Handle("POST /api/usenet/grab", authMW(http.HandlerFunc(s.usenetGrab)))
}

func (s *Server) resolveIndexer(ctx context.Context, userID string, plan plans.Plan) (*indexer.Client, bool) {
	if idx, err := s.Users.GetUsenetIndexer(ctx, userID); err == nil {
		return indexer.NewClient(idx.URL, idx.APIKey), true
	}
	if s.IndexerURL != "" && plan.SystemUsenet {
		return indexer.NewClient(s.IndexerURL, s.IndexerKey), true
	}
	return nil, false
}

func (s *Server) setIndexer(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	var req struct {
		URL    string `json:"url"`
		APIKey string `json:"api_key"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil || req.URL == "" || req.APIKey == "" {
		web.WriteError(w, 400, "url and api_key required")
		return
	}
	url := normalizeIndexerURL(req.URL)
	if err := indexer.ValidateURL(url); err != nil {
		web.WriteError(w, 400, "invalid indexer URL")
		return
	}
	if err := s.Users.SaveUsenetIndexer(r.Context(), user.ID, &auth.UsenetIndexer{URL: url, APIKey: req.APIKey}); err != nil {
		web.WriteError(w, 500, "failed to save indexer")
		return
	}
	s.Users.AuditLog(r.Context(), user.ID, "usenet_indexer_added", "url="+url, clientIP(r))
	web.WriteJSON(w, 200, map[string]string{"status": "ok"})
}

func (s *Server) testIndexer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL    string `json:"url"`
		APIKey string `json:"api_key"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil || req.URL == "" || req.APIKey == "" {
		web.WriteError(w, 400, "url and api_key required")
		return
	}
	url := normalizeIndexerURL(req.URL)
	if indexer.ValidateURL(url) != nil {
		web.WriteError(w, 400, "invalid indexer URL")
		return
	}
	if _, err := indexer.NewClient(url, req.APIKey).SearchQuery("test", "2000"); err != nil {
		web.WriteError(w, 400, "indexer connection failed")
		return
	}
	web.WriteJSON(w, 200, map[string]any{"ok": true})
}

func (s *Server) getIndexer(w http.ResponseWriter, r *http.Request) {
	idx, err := s.Users.GetUsenetIndexer(r.Context(), middleware.GetUser(r).ID)
	if err != nil {
		web.WriteJSON(w, 200, map[string]any{"configured": false})
		return
	}
	web.WriteJSON(w, 200, map[string]any{"configured": true, "url": idx.URL})
}

func (s *Server) delIndexer(w http.ResponseWriter, r *http.Request) {
	s.Users.DeleteUsenetIndexer(r.Context(), middleware.GetUser(r).ID)
	web.WriteJSON(w, 200, map[string]string{"status": "ok"})
}

func (s *Server) usenetSearch(w http.ResponseWriter, r *http.Request) {
	client, ok := s.resolveIndexer(r.Context(), middleware.GetUser(r).ID, middleware.GetPlan(r))
	if !ok {
		web.WriteError(w, 400, "configure a usenet indexer first")
		return
	}
	q := r.URL.Query()
	imdb, query := q.Get("imdb"), q.Get("q")
	season, _ := strconv.Atoi(q.Get("season"))
	episode, _ := strconv.Atoi(q.Get("ep"))

	var results []indexer.Result
	var err error
	switch {
	case imdb != "" && season > 0 && episode > 0:
		results, err = client.SearchTV(imdb, season, episode)
	case imdb != "":
		results, err = client.SearchMovie(imdb)
	case query != "":
		results, err = client.SearchQuery(query, q.Get("cat"))
	default:
		web.WriteError(w, 400, "provide imdb, q, or imdb+season+ep")
		return
	}
	if err != nil {
		web.WriteError(w, 502, "indexer search failed")
		return
	}
	if results == nil {
		results = []indexer.Result{}
	}
	web.WriteJSON(w, 200, results)
}

func (s *Server) usenetGrab(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	plan := middleware.GetPlan(r)
	client, ok := s.resolveIndexer(r.Context(), user.ID, plan)
	if !ok {
		web.WriteError(w, 400, "configure a usenet indexer first")
		return
	}
	var req struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil || req.ID == "" {
		web.WriteError(w, 400, "id required")
		return
	}
	nzbData, err := client.DownloadNZB(&indexer.Result{ID: req.ID})
	if err != nil {
		web.WriteError(w, 502, "failed to download NZB")
		return
	}
	s.ingestNZB(w, r, user, plan, nzbData, req.Title)
}

func normalizeIndexerURL(raw string) string {
	u := strings.TrimRight(strings.TrimSpace(raw), "/")
	u = strings.TrimSuffix(u, "/api")
	return strings.TrimRight(u, "/")
}
