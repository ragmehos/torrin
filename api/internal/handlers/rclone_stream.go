package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/torrin-app/torrin/api/internal/web"
	"github.com/torrin-app/torrin/shared/manifest"
	"github.com/torrin-app/torrin/shared/rclonerc"
)

func (s *Server) registerRcloneStreamRoutes(mux *http.ServeMux) {
	if s.RClone == nil {
		return
	}
	mux.HandleFunc("GET /api/rclone-stream/{job}/{idx}", s.rcloneStream)
}

func (s *Server) rcloneStream(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("job")
	idx, _ := strconv.Atoi(r.PathValue("idx"))
	exp, _ := strconv.ParseInt(r.URL.Query().Get("expires"), 10, 64)
	sig := r.URL.Query().Get("sig")
	if exp < time.Now().Unix() || !hmac.Equal([]byte(sig), []byte(rcloneStreamSig(s.SignKey, jobID, idx, exp))) {
		web.WriteError(w, 403, "invalid or expired link")
		return
	}

	job, err := s.Jobs.Get(r.Context(), jobID)
	if err != nil || job == nil || idx < 0 || idx >= len(job.Files) {
		web.WriteError(w, 404, "not found")
		return
	}
	creds, err := s.Users.GetStorageCreds(r.Context(), job.UserID)
	if err != nil || creds == nil || !creds.Enabled || !creds.IsRclone() {
		web.WriteError(w, 404, "not found")
		return
	}

	var params map[string]string
	json.Unmarshal([]byte(creds.ConfigJSON), &params)
	remote := rclonerc.UserRemoteName(job.UserID)
	sctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	s.RClone.CreateRemote(sctx, remote, creds.Backend, params, true)
	cancel()

	path := creds.Prefix + manifest.Key(job.InfoHash, idx, job.Files[idx].Name)
	proxyStream(w, r, fmt.Sprintf("%s/[%s:]/%s", s.RClone.Base(), remote, path))
}

func rcloneStreamSig(key []byte, jobID string, idx int, expires int64) string {
	mac := hmac.New(sha256.New, key)
	fmt.Fprintf(mac, "%s/%d:%d", jobID, idx, expires)
	return hex.EncodeToString(mac.Sum(nil))
}

func proxyStream(w http.ResponseWriter, r *http.Request, upstream string) {
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, upstream, nil)
	if err != nil {
		web.WriteError(w, 500, "stream error")
		return
	}
	if rng := r.Header.Get("Range"); rng != "" {
		req.Header.Set("Range", rng)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		web.WriteError(w, 502, "storage unavailable")
		return
	}
	defer resp.Body.Close()
	for _, h := range []string{"Content-Type", "Content-Length", "Content-Range", "Accept-Ranges", "Last-Modified", "ETag"} {
		if v := resp.Header.Get(h); v != "" {
			w.Header().Set(h, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
