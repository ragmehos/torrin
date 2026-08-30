package usenet

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"testing"

	"github.com/madcowfred/yencode"
	"github.com/torrin-app/torrin/shared/usenet/decoder"
	"github.com/torrin-app/torrin/shared/usenet/nzb"
)

type fakeFetcher struct {
	articles map[string][]byte
	err      error
	calls    []string
	groups   []string
}

func (f *fakeFetcher) Fetch(_ context.Context, id, group string) ([]byte, error) {
	f.calls = append(f.calls, id)
	f.groups = append(f.groups, group)
	if f.err != nil {
		return nil, f.err
	}
	return f.articles[id], nil
}

func article(data []byte, part, total int, begin int64) []byte {
	var body bytes.Buffer
	yencode.Encode(data, &body)
	return []byte(fmt.Sprintf("=ybegin part=%d total=%d line=128 size=%d name=x.bin\n=ypart begin=%d end=%d\n%s=yend size=%d part=%d pcrc32=%08X\n",
		part, total, len(data), begin+1, begin+int64(len(data)), body.String(), len(data), part, crc32.ChecksumIEEE(data)))
}

func testFile(parts ...[]byte) (nzb.File, map[string][]byte) {
	f := nzb.File{Groups: []string{"alt.test"}}
	articles := map[string][]byte{}
	var begin int64
	for i, part := range parts {
		id := fmt.Sprintf("part-%d", i+1)
		f.Segments = append(f.Segments, nzb.Segment{MessageID: id, Number: i + 1, Bytes: int64(len(part))})
		articles[id] = article(part, i+1, len(parts), begin)
		begin += int64(len(part))
	}
	return f, articles
}

func TestReaderLazySeekAndSegmentCache(t *testing.T) {
	file, articles := testFile([]byte("abcdef"), []byte("ghijkl"))
	fetch := &fakeFetcher{articles: articles}
	r, err := NewReader(context.Background(), file, fetch)
	if err != nil {
		t.Fatal(err)
	}
	if len(fetch.calls) != 0 {
		t.Fatal("constructor fetched an article")
	}

	buf := make([]byte, 4)
	if n, err := r.ReadAt(buf, 1); err != nil || n != 4 || string(buf) != "bcde" {
		t.Fatalf("first read: n=%d err=%v data=%q", n, err, buf)
	}
	if n, err := r.ReadAt(buf[:2], 3); err != nil || n != 2 || string(buf[:2]) != "de" {
		t.Fatalf("cached read: n=%d err=%v data=%q", n, err, buf[:2])
	}
	if len(fetch.calls) != 1 {
		t.Fatalf("same segment fetched %d times", len(fetch.calls))
	}

	buf = make([]byte, 6)
	if n, err := r.ReadAt(buf, 4); err != nil || n != 6 || string(buf) != "efghij" {
		t.Fatalf("cross read: n=%d err=%v data=%q", n, err, buf)
	}
	if got := fmt.Sprint(fetch.calls); got != "[part-1 part-2]" {
		t.Fatalf("calls = %s", got)
	}
	if fetch.groups[len(fetch.groups)-1] != "alt.test" {
		t.Fatalf("group = %q", fetch.groups[len(fetch.groups)-1])
	}
}

func TestReaderEOFAndInvalidInput(t *testing.T) {
	file, articles := testFile([]byte("abc"))
	r, _ := NewReader(context.Background(), file, &fakeFetcher{articles: articles})
	buf := make([]byte, 5)
	if n, err := r.ReadAt(buf, 1); n != 2 || !errors.Is(err, io.EOF) || string(buf[:n]) != "bc" {
		t.Fatalf("tail read: n=%d err=%v data=%q", n, err, buf[:n])
	}
	if n, err := r.ReadAt(buf, 3); n != 0 || !errors.Is(err, io.EOF) {
		t.Fatalf("eof read: n=%d err=%v", n, err)
	}
	if _, err := r.ReadAt(buf, -1); err == nil {
		t.Fatal("negative offset accepted")
	}
	if _, err := NewReader(context.Background(), nzb.File{}, &fakeFetcher{}); err == nil {
		t.Fatal("empty file accepted")
	}
}

func TestReaderPropagatesFetchAndDecodeErrors(t *testing.T) {
	file, articles := testFile([]byte("abc"))
	t.Run("fetch", func(t *testing.T) {
		r, _ := NewReader(context.Background(), file, &fakeFetcher{err: errors.New("offline")})
		if _, err := r.ReadAt(make([]byte, 1), 0); err == nil || !bytes.Contains([]byte(err.Error()), []byte("offline")) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("yenc", func(t *testing.T) {
		r, _ := NewReader(context.Background(), file, &fakeFetcher{articles: map[string][]byte{"part-1": []byte("bad")}})
		if _, err := r.ReadAt(make([]byte, 1), 0); err == nil {
			t.Fatal("invalid yenc accepted")
		}
	})
	t.Run("checksum", func(t *testing.T) {
		bad := append([]byte(nil), articles["part-1"]...)
		bad = bytes.Replace(bad, []byte(fmt.Sprintf("%08X", crc32.ChecksumIEEE([]byte("abc")))), []byte("DEADBEEF"), 1)
		r, _ := NewReader(context.Background(), file, &fakeFetcher{articles: map[string][]byte{"part-1": bad}})
		_, err := r.ReadAt(make([]byte, 1), 0)
		if !errors.Is(err, decoder.ErrChecksumMismatch) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("size", func(t *testing.T) {
		wrong := file
		wrong.Segments = append([]nzb.Segment(nil), file.Segments...)
		wrong.Segments[0].Bytes++
		r, _ := NewReader(context.Background(), wrong, &fakeFetcher{articles: articles})
		if _, err := r.ReadAt(make([]byte, 1), 0); err == nil {
			t.Fatal("decoded size mismatch accepted")
		}
	})
}
