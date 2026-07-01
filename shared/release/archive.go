package release

import (
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var Hosts = []string{"rapidgator", "nitroflare", "ddownload", "fikper"}

var sizeUnits = map[string]int64{"KB": 1 << 10, "MB": 1 << 20, "GB": 1 << 30, "TB": 1 << 40}

func ParseSize(s string) int64 {
	f := strings.Fields(s)
	if len(f) != 2 {
		return 0
	}
	n, err := strconv.ParseFloat(f[0], 64)
	if err != nil {
		return 0
	}
	return int64(n * float64(sizeUnits[strings.ToUpper(f[1])]))
}

func SplitTitle(text string) (title, size string) {
	if i := strings.LastIndex(text, "–"); i >= 0 {
		return strings.TrimSpace(text[:i]), strings.TrimSpace(text[i+len("–"):])
	}
	return text, ""
}

func BestArchive(doc *goquery.Document) []string {
	for _, host := range Hosts {
		if links := collectHost(doc, host); len(links) > 0 {
			return links
		}
	}
	return nil
}

func collectHost(doc *goquery.Document, host string) []string {
	seen := map[string]string{}
	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		if !strings.Contains(strings.ToLower(href), host) {
			return
		}
		clean := strings.TrimSuffix(href, ".html")
		name := strings.ToLower(path.Base(clean))
		if !isReleaseFile(name) || isSample(name) {
			return
		}
		if _, dup := seen[name]; !dup {
			seen[name] = clean
		}
	})
	names := make([]string, 0, len(seen))
	for n := range seen {
		names = append(names, n)
	}
	sort.Slice(names, func(i, j int) bool { return partNum(names[i]) < partNum(names[j]) })
	out := make([]string, len(names))
	for i, n := range names {
		out[i] = seen[n]
	}
	return out
}

func isReleaseFile(name string) bool {
	return strings.HasSuffix(name, ".rar") || strings.HasSuffix(name, ".mkv") || strings.HasSuffix(name, ".mp4")
}

func isSample(name string) bool {
	return strings.Contains(name, "sample") || strings.Contains(name, "proof")
}

func partNum(name string) int {
	i := strings.Index(name, ".part")
	if i < 0 {
		return 0
	}
	digits := ""
	for _, ch := range name[i+5:] {
		if ch < '0' || ch > '9' {
			break
		}
		digits += string(ch)
	}
	n, _ := strconv.Atoi(digits)
	return n
}
