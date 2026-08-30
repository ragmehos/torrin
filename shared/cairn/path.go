package cairn

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

func StreamPath(infoHash string, fileIndex int, name string) string {
	return fmt.Sprintf("%s/cairn/%d/%s", infoHash, fileIndex, filepath.Base(name))
}

func ParseStreamPath(path string) (infoHash string, fileIndex int, name string, ok bool) {
	parts := strings.Split(path, "/")
	if len(parts) != 4 || len(parts[0]) != 40 || parts[1] != "cairn" || parts[3] == "" {
		return "", 0, "", false
	}
	for _, c := range parts[0] {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return "", 0, "", false
		}
	}
	idx, err := strconv.Atoi(parts[2])
	if err != nil || idx < 0 {
		return "", 0, "", false
	}
	return parts[0], idx, parts[3], true
}
