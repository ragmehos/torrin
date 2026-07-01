package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"sync"

	tnp "github.com/torrin-app/torrent-name-parser"

	"github.com/torrin-app/torrin/api/internal/web"
	"github.com/torrin-app/torrin/shared/jobs"
	"github.com/torrin-app/torrin/shared/release"
	"github.com/torrin-app/torrin/shared/scenerls"
)

var (
	srOnce   sync.Once
	srClient *scenerls.Client
)

func scenerlsClient() *scenerls.Client {
	srOnce.Do(func() { srClient = scenerls.NewClient() })
	return srClient
}

func (s *Server) registerScenerlsRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
	mux.Handle("GET /api/scenerls/search", authMW(http.HandlerFunc(s.scenerlsSearch)))
	mux.Handle("POST /api/scenerls/grab", authMW(http.HandlerFunc(s.scenerlsGrab)))
}

func (s *Server) scenerlsSearch(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	if title == "" {
		web.WriteError(w, 400, "title required")
		return
	}
	var (
		results []release.Result
		err     error
	)
	season, _ := strconv.Atoi(r.URL.Query().Get("season"))
	if season > 0 {
		episode, _ := strconv.Atoi(r.URL.Query().Get("episode"))
		results, err = scenerlsClient().SearchTV(r.Context(), title, season, episode)
	} else {
		year, _ := strconv.Atoi(r.URL.Query().Get("year"))
		results, err = scenerlsClient().SearchMovie(r.Context(), title, year)
	}
	if err != nil {
		web.WriteError(w, 502, "scenerls search failed")
		return
	}
	s.writeReleaseResults(w, r, matchReleaseTitle(titleWants(r), results), jobs.SourceScenerls)
}

func matchReleaseTitle(wants []string, results []release.Result) []release.Result {
	norms := make(map[string]bool)
	for _, w := range wants {
		if n := jobs.NormTitle(w); n != "" {
			norms[n] = true
		}
	}
	if len(norms) == 0 {
		return results
	}
	out := make([]release.Result, 0, len(results))
	for _, res := range results {
		info, err := tnp.ParseName(strings.ReplaceAll(res.Title, ".", " "))
		if err == nil && norms[jobs.NormTitle(info.Title)] {
			out = append(out, res)
		}
	}
	return out
}

func titleWants(r *http.Request) []string {
	return append([]string{r.URL.Query().Get("title")}, r.URL.Query()["alias"]...)
}

func (s *Server) scenerlsGrab(w http.ResponseWriter, r *http.Request) {
	s.releaseGrab(w, r, jobs.SourceScenerls, "scene-rls requires a paid plan")
}
