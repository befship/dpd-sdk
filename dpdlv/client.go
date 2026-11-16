package dpdlv

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	eserviss "github.com/befship/dpd-sdk/internal/generated/eserviss"
)

const (
	// EnvSandbox is the DPD eServiss sandbox environment.
	EnvSandbox = "sandbox"
	// EnvProduction is the DPD eServiss production environment.
	EnvProduction = "production"

	// Market codes for Baltics eServiss hosts (17TRACK: LV=100809, LT=100810, EE=100808).
	MarketLV = "lv"
	MarketLT = "lt"
	MarketEE = "ee"
)

// Config holds DPD Baltics credentials and runtime options.
type Config struct {
	Username    string
	Password    string
	Token       string // optional static Bearer; skips /auth/tokens
	Environment string // sandbox (default) | production
	Market      string // lv (default) | lt | ee — production host selection
	BaseURL     string // optional override (trims trailing slash)

	HTTPClient *http.Client
	Timeout    time.Duration // when HTTPClient is nil; default 60s
	MaxRetries int           // HTTP retries on 429/5xx; default 2
	Logger     Logger
}

// Logger is a minimal structured logger hook.
type Logger interface {
	Debugf(format string, args ...any)
	Errorf(format string, args ...any)
}

// Creds is a compact credential bundle (used by adapters).
type Creds struct {
	Username    string
	Password    string
	Token       string
	Environment string
	Market      string // lv | lt | ee
	BaseURL     string
}

// Client is the hand-written DPD LV facade (token cache, Basic/Bearer, typed API).
type Client struct {
	cfg   Config
	http  *http.Client
	cache *tokenCache
}

// New builds a Client.
func New(cfg Config) *Client {
	if cfg.Environment == "" {
		cfg.Environment = EnvSandbox
	}
	switch strings.ToLower(cfg.Environment) {
	case "test", "cie", EnvSandbox:
		cfg.Environment = EnvSandbox
	case "prod", EnvProduction:
		cfg.Environment = EnvProduction
	}
	cfg.Market = NormalizeMarket(cfg.Market)
	if cfg.MaxRetries < 0 {
		cfg.MaxRetries = 0
	} else if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 2
	}
	hc := cfg.HTTPClient
	if hc == nil {
		timeout := cfg.Timeout
		if timeout <= 0 {
			timeout = 60 * time.Second
		}
		hc = &http.Client{Timeout: timeout}
	}
	c := &Client{cfg: cfg, http: hc, cache: newTokenCache()}
	c.ensureRetry()
	return c
}

// NormalizeMarket maps aliases to lv|lt|ee (default lv).
func NormalizeMarket(market string) string {
	switch strings.ToLower(strings.TrimSpace(market)) {
	case MarketLT, "lithuania":
		return MarketLT
	case MarketEE, "estonia":
		return MarketEE
	default:
		return MarketLV
	}
}

// Environment returns the configured environment name.
func (c *Client) Environment() string { return c.cfg.Environment }

// HTTPClient returns the underlying HTTP client.
func (c *Client) HTTPClient() *http.Client { return c.http }

// BaseURL returns the DPD eServiss API base for the given environment / override.
// Prefer BaseURLFor when selecting LT/EE production hosts.
func BaseURL(env string, override ...string) string {
	return BaseURLFor(env, MarketLV, override...)
}

// BaseURLFor resolves eServiss host by environment + Baltics market.
// Sandbox is always the LV sandbox host; production uses LV/LT/EE country portals.
func BaseURLFor(env, market string, override ...string) string {
	if len(override) > 0 && strings.TrimSpace(override[0]) != "" {
		return strings.TrimRight(strings.TrimSpace(override[0]), "/")
	}
	market = NormalizeMarket(market)
	switch strings.ToLower(strings.TrimSpace(env)) {
	case EnvProduction, "prod":
		switch market {
		case MarketLT:
			return "https://esiunta.dpd.lt/api/v1"
		case MarketEE:
			return "https://telli.dpd.ee/api/v1"
		default:
			return "https://eserviss.dpd.lv/api/v1"
		}
	default:
		return "https://sandbox-eserviss.dpd.lv/api/v1"
	}
}

// BaseURLFromCreds resolves the API base from a Creds bundle.
func BaseURLFromCreds(cr Creds) string {
	return BaseURLFor(cr.Environment, cr.Market, cr.BaseURL)
}

// BasicAuth returns a RequestEditorFn that sets Basic Authorization + Accept JSON.
func BasicAuth(username, password string) RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(username+":"+password)))
		if req.Header.Get("Accept") == "" {
			req.Header.Set("Accept", "application/json")
		}
		return nil
	}
}

// BearerAuth returns a RequestEditorFn that sets Bearer Authorization + Accept JSON.
func BearerAuth(token string) RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		if req.Header.Get("Accept") == "" {
			req.Header.Set("Accept", "application/json")
		}
		return nil
	}
}

// Raw builds an unauthenticated generated client for env (or cfg BaseURL / Market).
func (c *Client) Raw(env string, baseURL ...string) (*ClientWithResponses, error) {
	override := ""
	if len(baseURL) > 0 {
		override = baseURL[0]
	}
	return c.rawFor(env, c.cfg.Market, override)
}

func (c *Client) rawFor(env, market, override string) (*ClientWithResponses, error) {
	hc := c.http
	if hc == nil {
		hc = &http.Client{Timeout: 60 * time.Second}
	}
	if strings.TrimSpace(override) == "" {
		override = c.cfg.BaseURL
	}
	return eserviss.NewClientWithResponses(BaseURLFor(env, market, override), eserviss.WithHTTPClient(hc))
}

// AccessToken returns a static token or a cached/freshly fetched auth token.
func (c *Client) AccessToken(ctx context.Context, cr Creds) (string, error) {
	if cr.Token != "" {
		return cr.Token, nil
	}
	if cr.Username == "" || cr.Password == "" {
		if c.cfg.Token != "" {
			return c.cfg.Token, nil
		}
		if c.cfg.Username != "" && c.cfg.Password != "" {
			cr.Username = c.cfg.Username
			cr.Password = c.cfg.Password
		}
	}
	env := cr.Environment
	if env == "" {
		env = c.cfg.Environment
	}
	market := cr.Market
	if market == "" {
		market = c.cfg.Market
	}
	market = NormalizeMarket(market)
	key := env + ":" + market + ":" + cr.Username
	if tok, ok := c.cache.get(key); ok {
		return tok, nil
	}
	rc, err := c.rawFor(env, market, cr.BaseURL)
	if err != nil {
		return "", err
	}
	name := "befrate"
	ttl := int32(86400)
	resp, err := rc.CreateAuthTokensWithResponse(ctx, eserviss.TokenRequestDTO{
		Name: &name,
		Ttl:  &ttl,
	}, BasicAuth(cr.Username, cr.Password))
	if err != nil {
		return "", &APIError{Op: "auth", Message: "token request failed", Err: err}
	}
	if resp.StatusCode() >= 300 || resp.JSON200 == nil || resp.JSON200.Token == nil || *resp.JSON200.Token == "" {
		return "", &APIError{Op: "auth", StatusCode: resp.StatusCode(), Message: "token request rejected", Body: string(resp.Body)}
	}
	tok := *resp.JSON200.Token
	sec := int((23 * time.Hour).Seconds())
	if resp.JSON200.ValidUntil != nil {
		if t, err := time.Parse("2006-01-02 15:04", *resp.JSON200.ValidUntil); err == nil {
			sec = int(time.Until(t).Seconds())
			if sec < 60 {
				sec = 60
			}
		}
	}
	c.cache.set(key, tok, sec)
	return tok, nil
}

// Authed returns a generated client plus a Bearer RequestEditorFn.
func (c *Client) Authed(ctx context.Context, cr Creds) (*ClientWithResponses, RequestEditorFn, error) {
	tok, err := c.AccessToken(ctx, cr)
	if err != nil {
		return nil, nil, err
	}
	env := cr.Environment
	if env == "" {
		env = c.cfg.Environment
	}
	market := cr.Market
	if market == "" {
		market = c.cfg.Market
	}
	rc, err := c.rawFor(env, market, cr.BaseURL)
	if err != nil {
		return nil, nil, err
	}
	return rc, BearerAuth(tok), nil
}

// API returns an authenticated generated client using Config credentials.
func (c *Client) API(ctx context.Context) (*ClientWithResponses, RequestEditorFn, error) {
	return c.Authed(ctx, Creds{
		Username:    c.cfg.Username,
		Password:    c.cfg.Password,
		Token:       c.cfg.Token,
		Environment: c.cfg.Environment,
		Market:      c.cfg.Market,
		BaseURL:     c.cfg.BaseURL,
	})
}
