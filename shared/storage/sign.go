package storage

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"
)

func (c *Client) SignURL(path string, expiry time.Duration) string {
	return c.SignURLNode("", path, expiry)
}

func (c *Client) SignURLNode(node, path string, expiry time.Duration) string {
	expires := time.Now().Add(expiry).Unix()
	sig := c.sign(path, expires, "")
	return fmt.Sprintf("%s/%s?expires=%d&sig=%s", c.baseFor(node), escapePath(path), expires, sig)
}

func (c *Client) SignURLNodeUser(node, path, userID string, expiry time.Duration) string {
	if userID == "" {
		return c.SignURLNode(node, path, expiry)
	}
	expires := time.Now().Add(expiry).Unix()
	sig := c.sign(path, expires, userID)
	return fmt.Sprintf("%s/%s?expires=%d&u=%s&sig=%s",
		c.baseFor(node), escapePath(path), expires, url.QueryEscape(userID), sig)
}

func (c *Client) SignURLWithUser(path, userID string, expiry time.Duration) string {
	if userID == "" {
		return c.SignURL(path, expiry)
	}
	expires := time.Now().Add(expiry).Unix()
	sig := c.sign(path, expires, userID)
	return fmt.Sprintf("%s/%s?expires=%d&u=%s&sig=%s",
		c.publicURL, escapePath(path), expires, url.QueryEscape(userID), sig)
}

func (c *Client) Verify(path string, expires int64, userID, sig string) bool {
	if time.Now().Unix() > expires {
		return false
	}
	want, err := hex.DecodeString(sig)
	if err != nil {
		return false
	}
	got, _ := hex.DecodeString(c.sign(path, expires, userID))
	return hmac.Equal(want, got)
}

func (c *Client) sign(path string, expires int64, userID string) string {
	msg := fmt.Sprintf("%s:%d", path, expires)
	if userID != "" {
		msg = fmt.Sprintf("%s:%d:%s", path, expires, userID)
	}
	mac := hmac.New(sha256.New, c.signingKey)
	mac.Write([]byte(msg))
	return hex.EncodeToString(mac.Sum(nil))
}

func escapePath(path string) string {
	parts := strings.Split(path, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return strings.Join(parts, "/")
}
