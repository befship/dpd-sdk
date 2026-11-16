package dpduk

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type tokenJSON struct {
	Data *struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
		Expiry       int64  `json:"expiry"`
	} `json:"data"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	Expiry       int64  `json:"expiry"`
}

func parseTokenJSON(raw []byte) (TokenPair, error) {
	var tj tokenJSON
	if err := json.Unmarshal(raw, &tj); err != nil {
		return TokenPair{}, &APIError{Op: "auth", Message: "decode: " + err.Error(), Body: truncate(string(raw), 400), Auth: true}
	}
	out := TokenPair{
		AccessToken:  strings.TrimSpace(tj.AccessToken),
		RefreshToken: strings.TrimSpace(tj.RefreshToken),
		Expiry:       tj.Expiry,
	}
	if tj.Data != nil {
		if out.AccessToken == "" {
			out.AccessToken = strings.TrimSpace(tj.Data.AccessToken)
		}
		if out.RefreshToken == "" {
			out.RefreshToken = strings.TrimSpace(tj.Data.RefreshToken)
		}
		if out.Expiry == 0 {
			out.Expiry = tj.Data.Expiry
		}
	}
	if out.AccessToken == "" {
		return TokenPair{}, &APIError{Op: "auth", Message: "empty accessToken", Body: truncate(string(raw), 400), Auth: true}
	}
	return out, nil
}

// GetAccessToken exchanges api_key:api_secret for a JWT (POST /auth/token).
func (c *Client) GetAccessToken(ctx context.Context, cr Creds) (TokenPair, error) {
	if strings.TrimSpace(cr.APIKey) == "" || strings.TrimSpace(cr.APISecret) == "" {
		return TokenPair{}, invalidParam("api_key and api_secret are required")
	}
	rawURL := BaseURL(cr.Environment, cr.BaseURL) + PathAuthToken
	basic := base64.StdEncoding.EncodeToString([]byte(cr.APIKey + ":" + cr.APISecret))
	headers := map[string]string{
		"Authorization": "Basic " + basic,
		"Content-Type":  "application/json",
		"Accept":        "application/json",
	}
	raw, status, err := do(ctx, c.httpClient(), http.MethodPost, rawURL, headers, map[string]any{})
	if err != nil {
		return TokenPair{}, err
	}
	if status >= 300 {
		return TokenPair{}, &APIError{Op: "auth", StatusCode: status, Message: fmt.Sprintf("HTTP %d", status), Body: string(raw), Auth: true}
	}
	pair, err := parseTokenJSON(raw)
	if err != nil {
		return TokenPair{}, err
	}
	c.storeToken(cr, pair)
	return pair, nil
}

func (c *Client) storeToken(cr Creds, pair TokenPair) {
	exp := time.Now().Add(time.Duration(DefaultTokenTTLSeconds) * time.Second)
	if pair.Expiry > 0 {
		exp = time.Unix(pair.Expiry, 0)
	}
	c.mu.Lock()
	c.tokenCache()[tokenKey(cr)] = cachedToken{
		access:  pair.AccessToken,
		refresh: pair.RefreshToken,
		expiry:  exp,
	}
	c.mu.Unlock()
}

func (c *Client) ensureToken(ctx context.Context, cr Creds) (string, error) {
	c.mu.Lock()
	cached := c.tokenCache()[tokenKey(cr)]
	c.mu.Unlock()
	if cached.access != "" && time.Now().Before(cached.expiry.Add(-2*time.Minute)) {
		return cached.access, nil
	}
	pair, err := c.GetAccessToken(ctx, cr)
	if err != nil {
		return "", err
	}
	return pair.AccessToken, nil
}
