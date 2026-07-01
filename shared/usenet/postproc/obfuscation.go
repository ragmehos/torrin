package postproc

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func deobfuscateRars(dir string) {
	renameObfuscatedRars(dir)
	renameObfuscatedRarsByMagic(dir)
}

func renameObfuscatedRars(dir string) {
	entries, _ := os.ReadDir(dir)
	type rarPart struct{ name, partNum string }
	var parts []rarPart
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		lower := strings.ToLower(e.Name())
		idx := strings.Index(lower, ".part")
		if idx < 0 || !strings.HasSuffix(lower, ".rar") {
			continue
		}
		parts = append(parts, rarPart{name: e.Name(), partNum: lower[idx:]})
	}
	if len(parts) < 2 {
		return
	}
	base := func(n string) string {
		l := strings.ToLower(n)
		return l[:strings.Index(l, ".part")]
	}
	first := base(parts[0].name)
	for _, p := range parts[1:] {
		if base(p.name) != first {
			first = ""
			break
		}
	}
	if first != "" {
		return
	}
	slog.Info("renaming obfuscated RAR parts", "count", len(parts))
	for _, p := range parts {
		os.Rename(filepath.Join(dir, p.name), filepath.Join(dir, "archive"+p.partNum))
	}
}

func renameObfuscatedRarsByMagic(dir string) {
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasSuffix(strings.ToLower(e.Name()), ".rar") {
			return
		}
	}

	type rarFile struct {
		name, path string
		volNum     int
	}
	var rars []rarFile
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := strings.ToLower(e.Name())
		if strings.HasSuffix(name, ".par2") || strings.HasSuffix(name, ".nzb") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		if vol := rarVolumeNumber(path); vol >= 0 {
			rars = append(rars, rarFile{name: e.Name(), path: path, volNum: vol})
		}
	}
	if len(rars) == 0 {
		return
	}
	slog.Info("detected obfuscated RAR volumes by header", "count", len(rars))

	sort.Slice(rars, func(i, j int) bool {
		vi, vj := rars[i].volNum, rars[j].volNum
		if vi >= 0 && vj >= 0 && vi != vj {
			return vi < vj
		}
		return extractTrailingNumber(rars[i].name) < extractTrailingNumber(rars[j].name)
	})
	for i, r := range rars {
		newName := fmt.Sprintf("archive.part%03d.rar", i+1)
		if err := os.Rename(r.path, filepath.Join(dir, newName)); err != nil {
			slog.Warn("rename obfuscated rar failed", "from", r.name, "to", newName, "err", err)
		}
	}
}

func rarVolumeNumber(path string) int {
	f, err := os.Open(path)
	if err != nil {
		return -1
	}
	defer f.Close()
	buf := make([]byte, 64)
	n, _ := f.Read(buf)
	if n < 12 {
		return -1
	}
	// RAR5: Rar!\x1a\x07\x01\x00
	if string(buf[:4]) == "Rar!" && buf[4] == 0x1a && buf[5] == 0x07 && buf[6] == 0x01 && buf[7] == 0x00 {
		return rar5VolumeNumber(buf[8:n])
	}
	// RAR4: Rar!\x1a\x07\x00
	if string(buf[:4]) == "Rar!" && buf[4] == 0x1a && buf[5] == 0x07 && buf[6] == 0x00 {
		return rar4VolumeNumber(buf[7:n])
	}
	return -1
}

func rar5VolumeNumber(data []byte) int {
	if len(data) < 4 {
		return 0
	}
	off := 4
	_, n := readVint(data[off:])
	if n == 0 {
		return 0
	}
	off += n
	headerType, n := readVint(data[off:])
	if n == 0 || headerType != 1 {
		return 0
	}
	off += n
	headerFlags, n := readVint(data[off:])
	if n == 0 {
		return 0
	}
	off += n
	if headerFlags&0x0001 != 0 {
		_, n = readVint(data[off:])
		off += n
	}
	if headerFlags&0x0002 != 0 {
		_, n = readVint(data[off:])
		off += n
	}
	if off >= len(data) {
		return 0
	}
	archiveFlags, n := readVint(data[off:])
	if n == 0 {
		return 0
	}
	off += n
	if archiveFlags&0x0002 != 0 && off < len(data) {
		volNum, _ := readVint(data[off:])
		return int(volNum)
	}
	return 0
}

func rar4VolumeNumber(data []byte) int {
	if len(data) < 14 {
		return 0
	}
	mainHead := data[7:]
	if len(mainHead) < 7 {
		return 0
	}
	flags := uint16(mainHead[3]) | uint16(mainHead[4])<<8
	if flags&0x0001 == 0 {
		return 0
	}
	if flags&0x0100 != 0 {
		return 0
	}
	return 1
}

func readVint(data []byte) (uint64, int) {
	var val uint64
	for i, b := range data {
		if i >= 10 {
			return 0, 0
		}
		val |= uint64(b&0x7f) << (7 * uint(i))
		if b&0x80 == 0 {
			return val, i + 1
		}
	}
	return 0, 0
}

func extractTrailingNumber(name string) int {
	ext := filepath.Ext(name)
	if ext == "" {
		return 0
	}
	n := 0
	for _, c := range ext[1:] {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}
