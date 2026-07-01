package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func ValidateRD(ctx context.Context, key string) error { return validateKey(ctx, rdBase+"/user", key) }
func ValidateTB(ctx context.Context, key string) error {
	return validateKey(ctx, tbBase+"/user/me", key)
}

func ValidatePM(ctx context.Context, key string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pmBase+"/api/account/info", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not reach the provider — try again")
	}
	defer resp.Body.Close()
	var r struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil || r.Status != "success" {
		return fmt.Errorf("that key is invalid or expired")
	}
	return nil
}

func validateKey(ctx context.Context, url, key string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not reach the provider — try again")
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("that key is invalid or expired")
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("provider rejected the key (HTTP %d)", resp.StatusCode)
	}
	return nil
}
