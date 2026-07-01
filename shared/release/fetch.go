package release

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/torrin-app/torrin/shared/useragent"
)

type Result struct {
	Title     string `json:"title"`
	Size      string `json:"size"`
	SizeBytes int64  `json:"size_bytes"`
	PostURL   string `json:"post_url"`
}

func FetchDoc(ctx context.Context, client *http.Client, method, u string, body io.Reader) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", useragent.Default)
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("release http %d", resp.StatusCode)
	}
	return goquery.NewDocumentFromReader(resp.Body)
}
