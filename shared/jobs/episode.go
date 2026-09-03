package jobs

import (
	"slices"
	"strings"

	tnp "github.com/torrin-app/torrent-name-parser"
)

// MatchesEpisodeFile turns parser output into Torrin's playback decision. The
// parser owns filename syntax; Torrin owns request scoping and the rule that
// persisted job metadata is only a safe fallback for a genuine single file.
func MatchesEpisodeFile(j *Job, fileName string, season, episode int, single bool) bool {
	if season < 0 || episode <= 0 {
		return true
	}

	info, err := tnp.ParseName(strings.ReplaceAll(fileName, ".", " "))
	if err == nil && len(info.Episodes) > 0 {
		if !slices.Contains(info.Episodes, episode) {
			return false
		}
		if len(info.Seasons) == 0 {
			if j != nil && (j.Season > 0 || j.Episode > 0) {
				return j.Season == season
			}
			return season == 1
		}
		return slices.Contains(info.Seasons, season)
	}

	return single && j != nil && j.Season == season && j.Episode == episode
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
