package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const apiURL = "https://app.usesendi.com/api/emails"

type Client struct {
	apiKey string
	from   string
	http   *http.Client
}

func NewClient(apiKey, from string) *Client {
	return &Client{apiKey: apiKey, from: from, http: &http.Client{Timeout: 10 * time.Second}}
}

type sendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func (c *Client) Send(to, subject, html string) error {
	body, _ := json.Marshal(sendRequest{From: c.from, To: []string{to}, Subject: subject, HTML: html})
	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("sendi: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("sendi %d: %s", resp.StatusCode, b)
	}
	slog.Info("email sent", "to", to, "subject", subject)
	return nil
}

func (c *Client) SendWelcome(to, apiKey string) error {
	html := fmt.Sprintf(`<table cellpadding="0" cellspacing="0" border="0" width="100%%" style="max-width: 460px; margin: 0 auto; background: #111110;">
  <tr><td style="padding: 40px 0 24px; text-align: center;">
    <span style="font-family: -apple-system, sans-serif; font-size: 20px; font-weight: 600; color: #c8956c; letter-spacing: 0.02em;">torrin</span>
  </td></tr>
  <tr><td style="background: #191817; border-radius: 16px; padding: 36px 32px;">
    <p style="font-family: -apple-system, sans-serif; font-size: 15px; color: #d4d0cb; margin: 0 0 20px; line-height: 1.6;">you're in. here's your key:</p>
    <table cellpadding="0" cellspacing="0" border="0" width="100%%"><tr>
      <td style="background: #0f0e0d; border: 1px solid #2a2826; border-radius: 10px; padding: 14px 16px;">
        <code style="font-family: monospace; font-size: 13px; color: #c8956c; word-break: break-all;">%s</code>
      </td>
    </tr></table>
    <p style="font-family: -apple-system, sans-serif; font-size: 13px; color: #9b9590; margin: 14px 0 28px;">use it for Stremio, WebDAV, or the API.</p>
    <table cellpadding="0" cellspacing="0" border="0" width="100%%"><tr><td align="center">
      <a href="https://torrin.app/app/downloads" style="display: inline-block; background: #c8956c; color: #0f0e0d; font-family: -apple-system, sans-serif; font-size: 13px; font-weight: 600; text-decoration: none; padding: 10px 28px; border-radius: 8px;">open dashboard</a>
    </td></tr></table>
  </td></tr>
</table>`, apiKey)
	return c.Send(to, "your torrin API key", html)
}
