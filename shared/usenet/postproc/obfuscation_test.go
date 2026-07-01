package postproc

import (
	"os"
	"path/filepath"
	"testing"
)

// rar5File builds a minimal RAR5 main-archive header reporting the given volume num.
func rar5File(volNum byte) []byte {
	sig := []byte("Rar!\x1a\x07\x01\x00")
	hdr := []byte{
		0, 0, 0, 0, // 4-byte header CRC (skipped)
		0x05, // headerSize vint (consumed)
		0x01, // headerType = 1 (main archive)
		0x00, // headerFlags = 0
		0x02, // archiveFlags = volume-number present
		volNum,
	}
	return append(sig, hdr...)
}

func TestRarVolumeNumber(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "x")
	os.WriteFile(p, rar5File(3), 0o644)
	if got := rarVolumeNumber(p); got != 3 {
		t.Errorf("RAR5 vol = %d, want 3", got)
	}
	notrar := filepath.Join(dir, "y")
	os.WriteFile(notrar, []byte("this is not a rar file at all really"), 0o644)
	if got := rarVolumeNumber(notrar); got != -1 {
		t.Errorf("non-rar = %d, want -1", got)
	}
}

func TestRenameByMagic(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a8f3dat"), rar5File(2), 0o644) // volume 2
	os.WriteFile(filepath.Join(dir, "xQ9bin"), rar5File(1), 0o644)  // volume 1
	os.WriteFile(filepath.Join(dir, "junk.nzb"), []byte("nope"), 0o644)

	renameObfuscatedRarsByMagic(dir)

	if _, err := os.Stat(filepath.Join(dir, "archive.part001.rar")); err != nil {
		t.Error("expected archive.part001.rar")
	}
	if _, err := os.Stat(filepath.Join(dir, "archive.part002.rar")); err != nil {
		t.Error("expected archive.part002.rar")
	}
	// The volume-1 file (xQ9bin) must have become part001.
	if got := rarVolumeNumber(filepath.Join(dir, "archive.part001.rar")); got != 1 {
		t.Errorf("part001 should be the volNum=1 file, got vol %d", got)
	}
}

func TestRenameByMagicSkipsWhenRarPresent(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "movie.part01.rar"), rar5File(1), 0o644)
	os.WriteFile(filepath.Join(dir, "blob"), rar5File(2), 0o644)
	renameObfuscatedRarsByMagic(dir)
	if _, err := os.Stat(filepath.Join(dir, "blob")); err != nil {
		t.Error("should not touch files when a real .rar exists")
	}
}

func TestExtractTrailingNumber(t *testing.T) {
	if extractTrailingNumber("file.001") != 1 || extractTrailingNumber("x.123") != 123 {
		t.Error("trailing number parse")
	}
	if extractTrailingNumber("movie.mkv") != 0 {
		t.Error("non-numeric ext should be 0")
	}
}
