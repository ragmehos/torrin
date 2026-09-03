package jobs

import (
	"path"
	"regexp"
	"strconv"
	"strings"
)

var (
	sxePattern     = regexp.MustCompile(`(?i)(?:^|[^a-z0-9])s([0-9]{1,2})[ ._-]*e([0-9]{1,3})(?:[ ._-]*-[ ._-]*e?([0-9]{1,3}))?`)
	xPattern       = regexp.MustCompile(`(?i)(?:^|[^0-9])([0-9]{1,2})x([0-9]{1,3})(?:[ ._-]*-[ ._-]*([0-9]{1,3}))?`)
	extraEPattern  = regexp.MustCompile(`(?i)e([0-9]{1,3})`)
	seasonPattern  = regexp.MustCompile(`(?i)(?:^|[^a-z0-9])(?:s|season[ ._-]*)([0-9]{1,2})(?:[^0-9]|$)`)
	episodePattern = regexp.MustCompile(`(?i)(?:^|[^a-z0-9])(?:e|episode[ ._-]*)([0-9]{1,3})(?:[^0-9]|$)`)
)

type episodeScope struct {
	season   int
	episodes map[int]bool
}

// MatchesEpisodeFile requires an explicit season and episode match for files
// in a pack. Persisted job metadata is only trustworthy for a single-file job;
// a pack row represents the whole torrent, not the last episode requested.
func MatchesEpisodeFile(j *Job, fileName string, season, episode int, single bool) bool {
	if season < 0 || episode <= 0 {
		return true
	}

	if scope, ok := parseEpisodeScope(fileName); ok {
		return scope.season == season && scope.episodes[episode]
	}

	return single && j != nil && j.Season == season && j.Episode == episode
}

func parseEpisodeScope(fileName string) (episodeScope, bool) {
	withSlashes := strings.ReplaceAll(fileName, "\\", "/")
	baseName := path.Base(withSlashes)
	if scope, ok := parseSeasonEpisode(baseName, sxePattern); ok {
		return scope, true
	}
	if scope, ok := parseSeasonEpisode(baseName, xPattern); ok {
		return scope, true
	}

	// A few packs put only E03 in the basename and "Season 12" in a parent
	// directory. Use that fallback only when both pieces are explicit.
	epMatch := episodePattern.FindStringSubmatch(baseName)
	parent := path.Dir(withSlashes)
	seasonMatch := seasonPattern.FindStringSubmatch(parent)
	if len(epMatch) > 1 && len(seasonMatch) > 1 {
		ep, epOK := positiveInt(epMatch[1])
		packSeason, seasonOK := nonNegativeInt(seasonMatch[1])
		if epOK && seasonOK {
			return episodeScope{season: packSeason, episodes: map[int]bool{ep: true}}, true
		}
	}
	return episodeScope{}, false
}

func parseSeasonEpisode(name string, pattern *regexp.Regexp) (episodeScope, bool) {
	match := pattern.FindStringSubmatchIndex(name)
	if len(match) == 0 {
		return episodeScope{}, false
	}
	season, seasonOK := nonNegativeInt(name[match[2]:match[3]])
	first, episodeOK := positiveInt(name[match[4]:match[5]])
	if !seasonOK || !episodeOK {
		return episodeScope{}, false
	}
	episodes := map[int]bool{first: true}
	if match[6] >= 0 {
		last, ok := positiveInt(name[match[6]:match[7]])
		if ok {
			addEpisodeRange(episodes, first, last)
		}
	}
	// Also support adjacent multi-episode notation such as S12E01E02.
	for _, extra := range extraEPattern.FindAllStringSubmatch(name[match[5]:], -1) {
		if ep, ok := positiveInt(extra[1]); ok {
			episodes[ep] = true
		}
	}
	return episodeScope{season: season, episodes: episodes}, true
}

func addEpisodeRange(out map[int]bool, first, last int) {
	if last < first || last-first > 100 {
		out[last] = true
		return
	}
	for episode := first; episode <= last; episode++ {
		out[episode] = true
	}
}

func positiveInt(raw string) (int, bool) {
	n, err := strconv.Atoi(raw)
	return n, err == nil && n > 0
}

func nonNegativeInt(raw string) (int, bool) {
	n, err := strconv.Atoi(raw)
	return n, err == nil && n >= 0
}

// FilesForEpisode keeps movies/raw-hash requests unfiltered and preserves the
// torrent's original file indexes when selecting one episode from a pack.
func FilesForEpisode(j *Job, files []File, season, episode int) []File {
	indexed := false
	for _, f := range files {
		if f.Index > 0 {
			indexed = true
			break
		}
	}
	single := len(files) == 1
	out := make([]File, 0, len(files))
	for position, file := range files {
		if !indexed {
			file.Index = position
		}
		if MatchesEpisodeFile(j, file.Name, season, episode, single) {
			out = append(out, file)
		}
	}
	return out
}
