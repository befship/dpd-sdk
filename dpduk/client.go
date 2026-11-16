package dpduk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Client talks to DPD UK auth + shipping JSON endpoints.
type Client struct {
	HTTP *http.Client

	mu     sync.Mutex
	tokens map[string]cachedToken
}

type cachedToken struct {
	access  string
	refresh string
	expiry  time.Time
}

// Config holds optional HTTP settings for New.
type Config struct {
	HTTPClient *http.Client
	Timeout    time.Duration
}

// New returns a Client. Pass Config{} for defaults.
func New(cfg ...Config) *Client {
	c := &Client{tokens: make(map[string]cachedToken)}
	if len(cfg) > 0 && cfg[0].HTTPClient != nil {
		c.HTTP = cfg[0].HTTPClient
	} else {
		timeout := 60 * time.Second
		if len(cfg) > 0 && cfg[0].Timeout > 0 {
			timeout = cfg[0].Timeout
		}
		c.HTTP = &http.Client{Timeout: timeout}
	}
	return c
}

// NormalizeBaseURL trims trailing slashes.
func NormalizeBaseURL(base string) string {
	return strings.TrimRight(strings.TrimSpace(base), "/")
}

// BaseURL resolves production / sandbox host (or override).
func BaseURL(env, override string) string {
	if u := NormalizeBaseURL(override); u != "" {
		return u
	}
	switch strings.ToLower(strings.TrimSpace(env)) {
	case EnvSandbox, "test", "uat":
		return HostSandbox
	default:
		return HostProduction
	}
}

func (c *Client) httpClient() *http.Client {
	if c != nil && c.HTTP != nil {
		return c.HTTP
	}
	return http.DefaultClient
}

func (c *Client) tokenCache() map[string]cachedToken {
	if c.tokens == nil {
		c.tokens = make(map[string]cachedToken)
	}
	return c.tokens
}

func tokenKey(cr Creds) string {
	return BaseURL(cr.Environment, cr.BaseURL) + "|" + cr.APIKey
}

// InvalidateToken clears the cached JWT for cr.
func (c *Client) InvalidateToken(cr Creds) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.tokenCache(), tokenKey(cr))
}

func (c *Client) authedJSON(ctx context.Context, cr Creds, method, path string, payload, dest any) error {
	tok, err := c.ensureToken(ctx, cr)
	if err != nil {
		return err
	}
	rawURL := BaseURL(cr.Environment, cr.BaseURL) + path
	headers := map[string]string{
		"Authorization": "Bearer " + tok,
		"Client-Id":     cr.APIKey,
		"Content-Type":  "application/json",
		"Accept":        "application/json",
	}
	raw, status, err := do(ctx, c.httpClient(), method, rawURL, headers, payload)
	if err != nil {
		return err
	}
	if status == 401 {
		c.InvalidateToken(cr)
		tok, err = c.ensureToken(ctx, cr)
		if err != nil {
			return err
		}
		headers["Authorization"] = "Bearer " + tok
		raw, status, err = do(ctx, c.httpClient(), method, rawURL, headers, payload)
		if err != nil {
			return err
		}
	}
	if status >= 300 {
		var env apiEnvelope
		if json.Unmarshal(raw, &env) == nil && len(env.Error) > 0 {
			return mapAPIErrors(env.Error, raw)
		}
		return &APIError{Op: fmt.Sprintf("%s %s", method, path), StatusCode: status, Body: string(raw)}
	}
	return decodeData(raw, dest)
}
