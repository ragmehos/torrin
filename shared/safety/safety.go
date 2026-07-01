package safety

import (
	"context"
	"strings"
	"sync/atomic"
	"time"

	ahocorasick "github.com/petar-dambovaliev/aho-corasick"
)

type Verdict struct {
	Blocked bool
	Ban     bool
	Reason  string
}

type termSet struct {
	hard, soft *ahocorasick.AhoCorasick
}

var terms atomic.Pointer[termSet]

func SetTerms(hard, soft []string) {
	terms.Store(&termSet{hard: build(hard), soft: build(soft)})
}

func build(in []string) *ahocorasick.AhoCorasick {
	pats := clean(in)
	if len(pats) == 0 {
		return nil
	}
	b := ahocorasick.NewAhoCorasickBuilder(ahocorasick.Opts{
		AsciiCaseInsensitive: true,
		MatchOnlyWholeWords:  true,
		MatchKind:            ahocorasick.LeftMostLongestMatch,
	})
	ac := b.Build(pats)
	return &ac
}

func clean(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func hit(ac *ahocorasick.AhoCorasick, text string) bool {
	return ac != nil && len(ac.FindAll(text)) > 0
}

func Screen(texts ...string) Verdict {
	ts := terms.Load()
	var soft Verdict
	for _, t := range texts {
		if strings.Contains(strings.ToLower(t), ".onion") {
			return Verdict{Blocked: true, Ban: true, Reason: "blocked:onion"}
		}
		if ts == nil {
			continue
		}
		if hit(ts.hard, t) {
			return Verdict{Blocked: true, Ban: true, Reason: "blocked:term"}
		}
		if !soft.Blocked && hit(ts.soft, t) {
			soft = Verdict{Blocked: true, Ban: false, Reason: "flagged:term"}
		}
	}
	return soft
}

func Refresh(ctx context.Context, load func(context.Context) (hard, soft []string, err error), interval time.Duration) {
	apply := func() {
		if h, s, err := load(ctx); err == nil {
			SetTerms(h, s)
		}
	}
	apply()
	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				apply()
			}
		}
	}()
}
