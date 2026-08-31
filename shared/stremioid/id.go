package stremioid

import (
	"strconv"
	"strings"
)

// ID is the part of a Stremio content ID Torrin uses for routing. Series IDs
// have the form tt1234567:season:episode; raw info hashes remain supported for
// direct library lookups.
type ID struct {
	IMDBID   string
	InfoHash string
	Season   int
	Episode  int
}

func (id ID) IsEpisode() bool {
	return id.IMDBID != "" && id.Season > 0 && id.Episode > 0
}

func Parse(raw string) ID {
	parts := strings.Split(raw, ":")
	candidate := parts[0]
	if strings.HasPrefix(candidate, "tt") {
		id := ID{IMDBID: strings.TrimPrefix(candidate, "tt")}
		if len(parts) == 3 {
			season, seasonErr := strconv.Atoi(parts[1])
			episode, episodeErr := strconv.Atoi(parts[2])
			if seasonErr == nil && episodeErr == nil && season > 0 && episode > 0 {
				id.Season = season
				id.Episode = episode
			}
		}
		return id
	}
	if len(parts) == 1 && len(candidate) == 40 {
		return ID{InfoHash: strings.ToLower(candidate)}
	}
	return ID{}
}
