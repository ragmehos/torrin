package stremthru

import (
	"context"
	"encoding/json"
	"html"
	"net/http"

	"github.com/torrin-app/torrin/shared/auth"
	"github.com/torrin-app/torrin/shared/jobs"
	"github.com/torrin-app/torrin/shared/keyed"
	"github.com/torrin-app/torrin/shared/magnet"
	"github.com/torrin-app/torrin/shared/plans"
	"github.com/torrin-app/torrin/shared/stremioid"
)

func (h *Handler) listMagnets(w http.ResponseWriter, r *http.Request, user *auth.User) {
	userJobs, _ := jobs.ListAll(r.Context(), h.Jobs, user.ID)
	items := []map[string]any{}
	for _, j := range userJobs {
		items = append(items, h.magnetData(r.Context(), j))
	}
	stJSON(w, 200, map[string]any{"data": map[string]any{"items": items, "total_items": len(items)}})
}

func (h *Handler) getMagnet(w http.ResponseWriter, r *http.Request, user *auth.User) {
	job, err := h.Jobs.Get(r.Context(), r.PathValue("id"))
	if err != nil || job.UserID != user.ID {
		stError(w, 404, "not found")
		return
	}
	if job.Status == jobs.StatusComplete || job.Status == jobs.StatusSeeding {
		h.Jobs.RecordView(r.Context(), job.InfoHash, user.ID)
	}
	stJSON(w, 200, map[string]any{"data": h.magnetData(r.Context(), job)})
}

func (h *Handler) addMagnet(w http.ResponseWriter, r *http.Request, user *auth.User) {
	var req struct {
		Magnet string `json:"magnet"`
		Link   string `json:"link"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		stError(w, 400, "magnet required")
		return
	}
	if req.Magnet == "" {
		req.Magnet = req.Link
	}
	if req.Magnet == "" {
		stError(w, 400, "magnet required")
		return
	}
	req.Magnet = html.UnescapeString(req.Magnet)
	infoHash := extractHash(req.Magnet)
	if infoHash == "" {
		stError(w, 400, "invalid magnet")
		return
	}
	defer keyed.Lock(infoHash)()
	streamID := stremioid.Parse(r.URL.Query().Get("sid"))

	source, mag := jobs.SourceTorrent, req.Magnet
	hdTitle := ""
	var hdSize int64
	if pURL, t, src, sz := h.Jobs.ReleaseLink(r.Context(), infoHash); pURL != "" {
		if !plans.CanBYOK(user.PlanID) {
			stError(w, 403, "this release requires a paid plan")
			return
		}
		if jobs.Source(src) != jobs.SourceUsenet || h.canUsenet(r.Context(), user) {
			source, mag, hdTitle, hdSize = jobs.Source(src), pURL, t, sz
		}
	}

	cache, cached := h.cachedJobFiles(r.Context(), infoHash)
	plan, _ := plans.Get(user.PlanID)

	if streamID.IsEpisode() {
		if siblings, listErr := h.Jobs.ListByInfoHash(r.Context(), infoHash); listErr == nil {
			if reusable := reusableStreamJob(siblings, user.ID, streamID); reusable != nil {
				stJSON(w, 200, map[string]any{"data": h.magnetData(r.Context(), reusable)})
				return
			}
		}
	}

	existing, err := h.Jobs.GetByInfoHash(r.Context(), infoHash)
	if err == nil && existing != nil && existing.Status != jobs.StatusFailed && existing.Status != jobs.StatusEvicted {
		if existing.UserID == user.ID && sameStreamTarget(existing, streamID) {
			stJSON(w, 200, map[string]any{"data": h.magnetData(r.Context(), existing)})
			return
		}
		linked := &jobs.Job{
			UserID: user.ID, InfoHash: infoHash, Name: existing.Name, Magnet: mag,
			Source: source, Status: existing.Status, IMDBID: existing.IMDBID,
			Season: existing.Season, Episode: existing.Episode,
			Files: existing.Files, FileSize: existing.FileSize, Node: existing.Node,
		}
		applyStreamTarget(linked, streamID)
		activeLink := existing.Status.Active()
		if activeLink && !h.Slots.Acquire(r.Context(), user.ID, plan) {
			stError(w, 429, "slot limit reached")
			return
		}
		h.Jobs.Create(r.Context(), linked)
		if activeLink {
			h.Slots.Release(user.ID)
		}
		stJSON(w, 200, map[string]any{"data": h.magnetData(r.Context(), linked)})
		return
	}

	if !cached {
		if over, _ := h.Users.MonthlyQuotaExceeded(r.Context(), user.ID, plan.MonthlyIngestBytes); over {
			stError(w, 429, "monthly download limit reached, resets on the 1st")
			return
		}
	}
	if !cached && coldPullBlocked(r.Context(), h.Jobs, user.ID, plan.ColdPullsPerHour) {
		stError(w, 429, "hourly download limit reached, try later or upgrade")
		return
	}
	if !cached && !h.Slots.Acquire(r.Context(), user.ID, plan) {
		stError(w, 429, "slot limit reached")
		return
	}

	name := displayName(req.Magnet)
	if hdTitle != "" {
		name = hdTitle
	}
	job := &jobs.Job{
		UserID: user.ID, InfoHash: infoHash, Magnet: mag, Name: name, FileSize: hdSize,
		Source: source, IMDBID: streamID.IMDBID, Season: streamID.Season, Episode: streamID.Episode,
		Status: jobs.StatusPending, MaxBytes: plan.MaxTorrentBytes, Priority: plan.Priority,
	}
	if cached {
		job.Status = jobs.StatusComplete
		if cache.name != "" {
			job.Name = cache.name
		}
		job.FileSize, job.Files, job.Node = cache.size, cache.files, cache.node
	}
	h.Jobs.Create(r.Context(), job)
	if !cached {
		h.Slots.Release(user.ID)
		h.assign(job)
	}
	stJSON(w, 200, map[string]any{"data": h.magnetData(r.Context(), job)})
}

func (h *Handler) deleteMagnet(w http.ResponseWriter, r *http.Request, user *auth.User) {
	id := r.PathValue("id")
	job, err := h.Jobs.Get(r.Context(), id)
	if err != nil || job.UserID != user.ID {
		stError(w, 404, "not found")
		return
	}
	if job.Seed && job.Status == jobs.StatusSeeding {
		stError(w, 409, "seeding until it meets its ratio/time target")
		return
	}
	if job.Status == jobs.StatusComplete {
		if siblings, _ := h.Jobs.ListByInfoHash(r.Context(), job.InfoHash); len(siblings) <= 1 {
			job.UserID = "system"
			h.Jobs.Update(r.Context(), job)
			w.WriteHeader(204)
			return
		}
	}
	active := job.Status.Active()
	h.Jobs.Delete(r.Context(), id)
	if active && h.Qbit != nil {
		if siblings, _ := h.Jobs.ListByInfoHash(r.Context(), job.InfoHash); len(siblings) == 0 {
			h.Qbit.Login()
			h.Qbit.Delete(job.InfoHash)
		}
	}
	w.WriteHeader(204)
}

type coldPullChecker interface {
	ColdPullAllowed(ctx context.Context, userID string, perHour int) (bool, error)
}

func coldPullBlocked(ctx context.Context, c coldPullChecker, userID string, perHour int) bool {
	ok, err := c.ColdPullAllowed(ctx, userID, perHour)
	return err == nil && !ok
}

func displayName(m string) string { return magnet.DisplayName(m) }

func sameStreamTarget(j *jobs.Job, id stremioid.ID) bool {
	if id.IMDBID == "" {
		return true
	}
	if j.IMDBID != "" && j.IMDBID != id.IMDBID {
		return false
	}
	if id.IsEpisode() {
		return j.IMDBID == id.IMDBID && j.Season == id.Season && j.Episode == id.Episode
	}
	return j.Season == 0 && j.Episode == 0
}

func reusableStreamJob(candidates []*jobs.Job, userID string, id stremioid.ID) *jobs.Job {
	for _, candidate := range candidates {
		if candidate == nil || candidate.UserID != userID || candidate.Status == jobs.StatusFailed || candidate.Status == jobs.StatusEvicted {
			continue
		}
		if sameStreamTarget(candidate, id) {
			return candidate
		}
	}
	return nil
}

func applyStreamTarget(j *jobs.Job, id stremioid.ID) {
	if id.IMDBID != "" {
		j.IMDBID = id.IMDBID
		j.Season = id.Season
		j.Episode = id.Episode
	}
}

func (h *Handler) canUsenet(ctx context.Context, user *auth.User) bool {
	plan, _ := plans.Get(user.PlanID)
	if plan.SystemUsenet {
		return true
	}
	_, err := h.Users.GetUsenetCreds(ctx, user.ID)
	return err == nil && plans.CanBYOK(plan.ID)
}
