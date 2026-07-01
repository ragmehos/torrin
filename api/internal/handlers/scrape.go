package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/torrin-app/torrin/api/internal/web"
)

func (s *Server) scrapeHashes(w http.ResponseWriter, r *http.Request) {
	if s.Scrape == nil {
		web.WriteJSON(w, 200, map[string]any{})
		return
	}
	var req struct {
		Hashes []string `json:"hashes"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		web.WriteError(w, 400, "invalid json")
		return
	}
	if len(req.Hashes) == 0 {
		web.WriteError(w, 400, "hashes required")
		return
	}
	web.WriteJSON(w, 200, s.Scrape.Check(r.Context(), req.Hashes))
}
