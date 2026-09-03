package stremioid

import (
	"encoding/hex"
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
	return id.IMDBID != "" && id.Season >= 0 && id.Episode > 0
}

func Parse(raw string) ID {
	parts := strings.Split(strings.TrimSpace(raw), ":")
	candidate := parts[0]
	if strings.HasPrefix(candidate, "tt") {
		imdbID := strings.TrimPrefix(candidate, "tt")
		if imdbID == "" || !decimal(imdbID) {
			return ID{}
		}
		if len(parts) == 1 {
			return ID{IMDBID: imdbID}
		}
		if len(parts) != 3 {
			return ID{}
		}
		season, seasonErr := strconv.Atoi(parts[1])
		episode, episodeErr := strconv.Atoi(parts[2])
		if seasonErr != nil || episodeErr != nil || season < 0 || episode <= 0 {
			return ID{}
		}
		return ID{IMDBID: imdbID, Season: season, Episode: episode}
	}
	if len(parts) == 1 && len(candidate) == 40 && validHex(candidate) {
		return ID{InfoHash: strings.ToLower(candidate)}
	}
	return ID{}
}

func decimal(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func validHex(s string) bool {
	_, err := hex.DecodeString(s)
	return err == nil
}
