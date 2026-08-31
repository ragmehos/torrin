package stremthru

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/torrin-app/torrin/shared/auth"
	"github.com/torrin-app/torrin/shared/crypto"
	"github.com/torrin-app/torrin/shared/jobs"
	"github.com/torrin-app/torrin/shared/manifest"
	"github.com/torrin-app/torrin/shared/plans"
	"github.com/torrin-app/torrin/shared/stremioid"
	"github.com/torrin-app/torrin/shared/usenet/nzb"
)

func TestReleaseLinkPaidGate(t *testing.T) {
	if plans.CanBYOK("free") {
		t.Error("free must not cache hoster/hdencode release-link content")
	}
	for _, p := range []string{"starter", "standard", "pro"} {
		if !plans.CanBYOK(p) {
			t.Errorf("%s should be allowed to cache release-link content", p)
		}
	}
}

type fakeStore struct {
	manifest []byte
	err      error
	missing  bool
}

func (f fakeStore) Has(context.Context, string) (bool, error) { return !f.missing, nil }
func (f fakeStore) GetBytes(context.Context, string) ([]byte, error) {
	return f.manifest, f.err
}
func (f fakeStore) Put(context.Context, string, io.Reader, string) error { return nil }
func (f fakeStore) SignURL(path string, _ time.Duration) string          { return "sign://" + path }
func (f fakeStore) SignURLNode(_, path string, _ time.Duration) string   { return "sign://" + path }
func (f fakeStore) SignURLNodeUser(_, path, userID string, _ time.Duration) string {
	return "sign://" + path + "?u=" + userID
}

func packJob() *jobs.Job {
	return &jobs.Job{
		InfoHash: "abc", Status: jobs.StatusSeeding,
		Files: []jobs.File{
			{Name: "Reborn.Rookie.S01.1080p/Reborn.Rookie.S01E01.mkv", Size: 100},
			{Name: "Reborn.Rookie.S01.1080p/Reborn.Rookie.S01E07.mkv", Size: 200},
		},
	}
}

func TestBuildFileEntries(t *testing.T) {
	h := &Handler{Deps: Deps{Store: fakeStore{}}}
	out := h.buildFileEntries("user-1", "abc", "box2", packJob().Files)
	if len(out) != 2 {
		t.Fatalf("got %d entries, want 2", len(out))
	}
	if out[0]["size"] != int64(100) || out[1]["size"] != int64(200) {
		t.Errorf("sizes wrong: %v %v", out[0]["size"], out[1]["size"])
	}
	if link, _ := out[1]["link"].(string); !strings.HasPrefix(link, "sign://") {
		t.Errorf("entry missing node-signed link: %v", out[1])
	}
}

func TestExtractHash(t *testing.T) {
	h := "0123456789abcdef0123456789abcdef01234567"
	if got := extractHash("magnet:?xt=urn:btih:" + h + "&dn=x"); got != h {
		t.Errorf("magnet = %q", got)
	}
	if got := extractHash("xt=urn:btih:" + h); got != h {
		t.Errorf("bare = %q", got)
	}
	if got := extractHash("magnet:?xt=urn:btih:tooshort"); got != "" {
		t.Errorf("invalid should be empty, got %q", got)
	}
	if got := extractHash("https://example.com/x"); got != "" {
		t.Errorf("non-magnet should be empty, got %q", got)
	}
}

func TestStStatus(t *testing.T) {
	cases := map[jobs.Status]string{
		jobs.StatusComplete:    "downloaded",
		jobs.StatusDownloading: "downloading",
		jobs.StatusProcessing:  "downloading",
		jobs.StatusPublishing:  "downloading",
		jobs.StatusSeeding:     "downloaded",
		jobs.StatusFailed:      "failed",
		jobs.StatusPending:     "queued",
		jobs.StatusQueued:      "queued",
	}
	for in, want := range cases {
		if got := stStatus(in); got != want {
			t.Errorf("stStatus(%s)=%q want %q", in, got, want)
		}
	}
}

type fakeColdPull struct {
	allowed bool
	err     error
}

func (f fakeColdPull) ColdPullAllowed(context.Context, string, int) (bool, error) {
	return f.allowed, f.err
}

func TestColdPullBlocked(t *testing.T) {
	if !coldPullBlocked(context.Background(), fakeColdPull{allowed: false}, "u", 15) {
		t.Error("over-limit user must be blocked")
	}
	if coldPullBlocked(context.Background(), fakeColdPull{allowed: true}, "u", 15) {
		t.Error("under-limit user must not be blocked")
	}
	if coldPullBlocked(context.Background(), fakeColdPull{allowed: false, err: context.DeadlineExceeded}, "u", 15) {
		t.Error("must fail open: a checker error must not block the add")
	}
}

func TestSameStreamTarget(t *testing.T) {
	j := &jobs.Job{IMDBID: "3121722", Season: 5, Episode: 1}
	if !sameStreamTarget(j, stremioid.Parse("tt3121722:5:1")) {
		t.Fatal("same episode should reuse the job")
	}
	if sameStreamTarget(j, stremioid.Parse("tt3121722:5:2")) {
		t.Fatal("another episode in the same pack needs a distinct linked job")
	}
	if sameStreamTarget(j, stremioid.Parse("tt9999999:5:1")) {
		t.Fatal("another title needs a distinct linked job")
	}
	if sameStreamTarget(j, stremioid.Parse("tt3121722")) {
		t.Fatal("movie requests must not inherit an episode-scoped job")
	}
	if !sameStreamTarget(&jobs.Job{IMDBID: "3121722"}, stremioid.Parse("tt3121722")) {
		t.Fatal("movie requests should reuse an unscoped job")
	}
	if !sameStreamTarget(j, stremioid.ID{}) {
		t.Fatal("requests without a Stremio target keep existing reuse behavior")
	}
}

func TestApplyStreamTarget(t *testing.T) {
	j := &jobs.Job{IMDBID: "old", Season: 12, Episode: 3}
	applyStreamTarget(j, stremioid.Parse("tt3121722:5:1"))
	if j.IMDBID != "3121722" || j.Season != 5 || j.Episode != 1 {
		t.Fatalf("target = imdb %q S%02dE%02d", j.IMDBID, j.Season, j.Episode)
	}
}

func TestApplyMovieTargetClearsEpisodeScope(t *testing.T) {
	j := &jobs.Job{IMDBID: "old", Season: 12, Episode: 3}
	applyStreamTarget(j, stremioid.Parse("tt0816692"))
	if j.IMDBID != "0816692" || j.Season != 0 || j.Episode != 0 {
		t.Fatalf("movie target = imdb %q S%02dE%02d", j.IMDBID, j.Season, j.Episode)
	}
}

func TestReusableStreamJobFindsExistingEpisode(t *testing.T) {
	want := &jobs.Job{ID: "wanted", UserID: "user", IMDBID: "3121722", Season: 5, Episode: 1, Status: jobs.StatusComplete}
	candidates := []*jobs.Job{
		{ID: "other-episode", UserID: "user", IMDBID: "3121722", Season: 5, Episode: 2, Status: jobs.StatusComplete},
		{ID: "other-user", UserID: "other", IMDBID: "3121722", Season: 5, Episode: 1, Status: jobs.StatusComplete},
		{ID: "failed", UserID: "user", IMDBID: "3121722", Season: 5, Episode: 1, Status: jobs.StatusFailed},
		want,
	}
	if got := reusableStreamJob(candidates, "user", stremioid.Parse("tt3121722:5:1")); got != want {
		t.Fatalf("reusable job = %+v, want %q", got, want.ID)
	}
}

func TestFileEntry(t *testing.T) {
	e := fileEntry(2, "the.100.s02e01.mkv", 3517219191, "https://beam/link", nil)
	if e["path"] != "/the.100.s02e01.mkv" {
		t.Errorf("path = %v, want /the.100.s02e01.mkv", e["path"])
	}
	for _, k := range []string{"index", "name", "path", "size", "link"} {
		if _, ok := e[k]; !ok {
			t.Errorf("missing key %q", k)
		}
	}
}

func TestDisplayName(t *testing.T) {
	m := "magnet:?xt=urn:btih:abcdef&dn=Big+Buck+Bunny+2008+1080p&tr=udp://x"
	if got := displayName(m); got != "Big Buck Bunny 2008 1080p" {
		t.Errorf("dn = %q", got)
	}
	if got := displayName("magnet:?xt=urn:btih:abcdef"); got != "" {
		t.Errorf("no dn should be empty, got %q", got)
	}
}

func TestMagnetDataUsesManifestBasenames(t *testing.T) {
	m := manifest.Manifest{
		InfoHash: "abc", Name: "Reborn.Rookie.S01.1080p",
		Files: []manifest.File{
			{FileName: "Reborn.Rookie.S01E01.mkv", FileSize: 100},
			{FileName: "Reborn.Rookie.S01E07.mkv", FileSize: 200},
		},
	}
	data, _ := m.Marshal()
	h := &Handler{Deps: Deps{Store: fakeStore{manifest: data}}}

	files, _ := h.magnetData(context.Background(), packJob())["files"].([]map[string]any)
	if len(files) != 2 {
		t.Fatalf("files = %d, want 2", len(files))
	}
	// name must be the manifest basename, not the folder-prefixed DB name
	if files[1]["name"] != "Reborn.Rookie.S01E07.mkv" {
		t.Errorf("name = %q, want basename", files[1]["name"])
	}
	// the R2 key in the link must use the basename too
	if link, _ := files[1]["link"].(string); !strings.Contains(link, "abc/file_1/Reborn.Rookie.S01E07.mkv") {
		t.Errorf("link = %q, want basename key", link)
	}
}

func TestMagnetDataFallsBackToJobFilesWithoutManifest(t *testing.T) {
	h := &Handler{Deps: Deps{Store: fakeStore{err: context.DeadlineExceeded}}}
	files, _ := h.magnetData(context.Background(), packJob())["files"].([]map[string]any)
	if len(files) != 2 {
		t.Fatalf("files = %d, want 2", len(files))
	}
	if files[1]["name"] != "Reborn.Rookie.S01.1080p/Reborn.Rookie.S01E07.mkv" {
		t.Errorf("fallback name = %q, want job file name", files[1]["name"])
	}
}

func TestMagnetDataFiltersSeasonPackToRequestedEpisode(t *testing.T) {
	m := manifest.Manifest{
		InfoHash: "abc", Name: "Paw.Patrol.S05.1080p",
		Files: []manifest.File{
			{FileName: "Paw.Patrol.S05E01.mkv", FileSize: 100},
			{FileName: "Paw.Patrol.S05E02.mkv", FileSize: 200},
			{FileName: "Paw.Patrol.S12E01.mkv", FileSize: 300},
		},
	}
	data, _ := m.Marshal()
	h := &Handler{Deps: Deps{Store: fakeStore{manifest: data}}}
	j := &jobs.Job{InfoHash: "abc", Status: jobs.StatusComplete, Season: 5, Episode: 2}

	files, _ := h.magnetData(context.Background(), j)["files"].([]map[string]any)
	if len(files) != 1 {
		t.Fatalf("files = %d, want one S05E02 file", len(files))
	}
	if files[0]["name"] != "Paw.Patrol.S05E02.mkv" {
		t.Fatalf("name = %v, want S05E02", files[0]["name"])
	}
	if files[0]["index"] != 1 {
		t.Fatalf("index = %v, want original manifest index 1", files[0]["index"])
	}
	if link, _ := files[0]["link"].(string); !strings.Contains(link, "abc/file_1/Paw.Patrol.S05E02.mkv") {
		t.Fatalf("link = %q, want original file_1 key", link)
	}
}

func TestMagnetDataIncludesMagnet(t *testing.T) {
	h := &Handler{Deps: Deps{Store: fakeStore{err: context.DeadlineExceeded}}}
	j := packJob()
	j.Magnet = "https://scene-rls.net/some-release/"
	d := h.magnetData(context.Background(), j)
	m, _ := d["magnet"].(string)
	if !strings.HasPrefix(m, "magnet:?xt=urn:btih:abc") {
		t.Errorf("magnet = %q, want a proper magnet URI (not the stored source)", m)
	}
	if _, ok := d["name"]; !ok {
		t.Error("name key missing")
	}
}

func TestCachedFilesReturnsManifestName(t *testing.T) {
	m := manifest.Manifest{
		InfoHash: "abc", Name: "Some.Movie.2020.1080p",
		Files: []manifest.File{{FileName: "movie.mkv", FileSize: 100}},
	}
	data, _ := m.Marshal()
	h := &Handler{Deps: Deps{Store: fakeStore{manifest: data}}}

	name, files, ok := h.cachedFiles(context.Background(), "user-1", "abc")
	if !ok || len(files) != 1 {
		t.Fatalf("ok=%v files=%d", ok, len(files))
	}
	if name != "Some.Movie.2020.1080p" {
		t.Errorf("name = %q, want manifest name", name)
	}
}

type fakeCairnRepository struct {
	hash string
	name string
	nzb  []byte
}

func (f fakeCairnRepository) GetCairnArchive(_ context.Context, hash string) (string, string, bool) {
	return "nzb/" + hash + ".nzb", f.name, hash == f.hash
}

func (f fakeCairnRepository) GetCairnNZB(_ context.Context, hash string) ([]byte, bool) {
	return f.nzb, hash == f.hash && len(f.nzb) > 0
}

func directCairnHandler(t *testing.T, hash string) *Handler {
	t.Helper()
	cipher, err := crypto.NewStream(strings.Repeat("ab", 32))
	if err != nil {
		t.Fatal(err)
	}
	plainSize := int64(100000)
	encSize, err := cipher.EncryptedSize(plainSize)
	if err != nil {
		t.Fatal(err)
	}
	nzbData := nzb.Generate([]nzb.OutFile{{Name: "movie.mkv", Group: "alt.test", Segments: []nzb.Segment{
		{MessageID: "part-1", Number: 1, Bytes: encSize},
	}}})
	return New(Deps{
		Store:      fakeStore{err: context.DeadlineExceeded, missing: true},
		Cairns:     fakeCairnRepository{hash: hash, name: "Cold Movie", nzb: nzbData},
		CairnStore: fakeStore{manifest: nzbData}, CairnCipher: cipher, CairnDirect: true,
	})
}

func TestCheckMagnetsReportsDirectCairnAsCached(t *testing.T) {
	hash := strings.Repeat("a", 40)
	h := directCairnHandler(t, hash)
	r := httptest.NewRequest("GET", "/v0/store/magnets/check?magnet="+hash, nil)
	w := httptest.NewRecorder()
	h.checkMagnets(w, r, &auth.User{ID: "user-1", PlanID: "standard"})
	if w.Code != 200 {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var response struct {
		Data struct {
			Items []struct {
				Status string           `json:"status"`
				Files  []map[string]any `json:"files"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Data.Items) != 1 || response.Data.Items[0].Status != "cached" || len(response.Data.Items[0].Files) != 1 {
		t.Fatalf("unexpected check response: %s", w.Body.String())
	}
	file := response.Data.Items[0].Files[0]
	link, _ := file["link"].(string)
	if !strings.Contains(link, hash+"/cairn/0/movie.mkv") || !strings.Contains(link, "u=user-1") || !strings.Contains(link, "enc=1") {
		t.Errorf("direct cairn link = %q", link)
	}
	if file["size"] != float64(100000) {
		t.Errorf("plaintext size = %v, want 100000", file["size"])
	}
}

func TestMagnetDataResolvesEvictedCairnDirectly(t *testing.T) {
	hash := strings.Repeat("b", 40)
	h := directCairnHandler(t, hash)
	data := h.magnetData(context.Background(), &jobs.Job{
		ID: "job-1", UserID: "user-2", InfoHash: hash, Name: "Evicted", Status: jobs.StatusEvicted,
	})
	if data["status"] != "downloaded" || data["name"] != "Cold Movie" {
		t.Fatalf("resolved item = %+v", data)
	}
	files, _ := data["files"].([]map[string]any)
	if len(files) != 1 {
		t.Fatalf("resolved files = %+v", data["files"])
	}
	link, _ := files[0]["link"].(string)
	if !strings.Contains(link, hash+"/cairn/0/movie.mkv") || !strings.Contains(link, "u=user-2") {
		t.Errorf("resolved link = %q", link)
	}
}

func TestMagnetDataFiltersCairnSeasonPackAndPreservesNZBIndex(t *testing.T) {
	hash := strings.Repeat("d", 40)
	nzbData := nzb.Generate([]nzb.OutFile{
		{Name: "Paw.Patrol.S05E01.mkv", Group: "alt.test", Segments: []nzb.Segment{{MessageID: "part-1", Number: 1, Bytes: 100}}},
		{Name: "Paw.Patrol.S05E02.mkv", Group: "alt.test", Segments: []nzb.Segment{{MessageID: "part-2", Number: 1, Bytes: 200}}},
		{Name: "Paw.Patrol.S12E01.mkv", Group: "alt.test", Segments: []nzb.Segment{{MessageID: "part-3", Number: 1, Bytes: 300}}},
	})
	h := New(Deps{
		Store:      fakeStore{err: context.DeadlineExceeded, missing: true},
		Cairns:     fakeCairnRepository{hash: hash, name: "Paw Patrol", nzb: nzbData},
		CairnStore: fakeStore{manifest: nzbData}, CairnDirect: true,
	})
	data := h.magnetData(context.Background(), &jobs.Job{
		ID: "job-1", UserID: "user-1", InfoHash: hash, Name: "Paw Patrol",
		Status: jobs.StatusEvicted, Season: 5, Episode: 2,
	})
	files, _ := data["files"].([]map[string]any)
	if len(files) != 1 {
		t.Fatalf("files = %+v, want one S05E02 Cairn file", files)
	}
	if files[0]["name"] != "Paw.Patrol.S05E02.mkv" || files[0]["index"] != 1 {
		t.Fatalf("file = %+v, want original NZB index 1", files[0])
	}
	link, _ := files[0]["link"].(string)
	if !strings.Contains(link, hash+"/cairn/1/Paw.Patrol.S05E02.mkv") || !strings.Contains(link, "u=user-1") {
		t.Fatalf("link = %q, want user-bound Cairn index 1", link)
	}
}

func TestWarmCacheWinsOverDirectCairn(t *testing.T) {
	hash := strings.Repeat("c", 40)
	h := directCairnHandler(t, hash)
	warmManifest, err := (manifest.Manifest{InfoHash: hash, Name: "Warm Movie", Files: []manifest.File{
		{FileName: "warm.mkv", DirectURL: "warm/object", FileSize: 1234},
	}}).Marshal()
	if err != nil {
		t.Fatal(err)
	}
	h.Store = fakeStore{manifest: warmManifest}
	name, files, ok := h.cachedFiles(context.Background(), "user-3", hash)
	if !ok || name != "Warm Movie" || len(files) != 1 {
		t.Fatalf("warm result: ok=%v name=%q files=%+v", ok, name, files)
	}
	link, _ := files[0]["link"].(string)
	if strings.Contains(link, "/cairn/") || !strings.Contains(link, "warm/object") {
		t.Errorf("warm cache did not win: %q", link)
	}
}
