package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/torrin-app/torrin/api/internal/middleware"
	"github.com/torrin-app/torrin/api/internal/web"
	"github.com/torrin-app/torrin/shared/jobs"
	"github.com/torrin-app/torrin/shared/release"
)

func (s *Server) writeReleaseResults(w http.ResponseWriter, r *http.Request, results []release.Result, source jobs.Source) {
	out := make([]map[string]any, len(results))
	for i, res := range results {
		ih := hosterInfoHash(res.PostURL)
		s.JobsPG.StoreReleaseLink(r.Context(), ih, res.PostURL, res.Title, string(source), res.SizeBytes)
		out[i] = map[string]any{
			"title":     res.Title,
			"size":      res.Size,
			"post_url":  res.PostURL,
			"info_hash": ih,
		}
	}
	web.WriteJSON(w, 200, out)
}

func (s *Server) releaseGrab(w http.ResponseWriter, r *http.Request, source jobs.Source, planMsg string) {
	user := middleware.GetUser(r)
	if middleware.GetPlan(r).ID == "free" {
		web.WriteError(w, 403, planMsg)
		return
	}
	var req struct {
		PostURL string `json:"post_url"`
		Title   string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PostURL == "" {
		web.WriteError(w, 400, "post_url required")
		return
	}
	if s.blockedBySafety(w, r, user.ID, req.Title) {
		return
	}
	s.submitMagnet(w, r, hosterInfoHash(req.PostURL), req.PostURL, req.Title, source, false)
}
