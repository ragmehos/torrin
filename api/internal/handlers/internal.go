package handlers

import "net/http"

func (s *Server) recordView(w http.ResponseWriter, r *http.Request) {
	if hash := r.PathValue("hash"); hash != "" {
		s.Jobs.RecordView(r.Context(), hash, "stream")
	}
	w.WriteHeader(http.StatusNoContent)
}
