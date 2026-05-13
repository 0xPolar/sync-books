package kavita

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Client is a Kavita API client. It transparently handles JWT exchange and
// refresh — pass an API key (from Kavita: Account → API Key) and a plugin
// name (any short identifier; shows up in Kavita's logs).
//
// Zero value isn't usable; construct with New.
type Client struct {
	baseURL    *url.URL
	apiKey     string
	pluginName string
	http       *http.Client

	mu           sync.RWMutex
	token        string
	refreshToken string
	userID       int
}

// New builds a Client. baseURL should be the full Kavita root, e.g.
// "https://kavita.aoaknode.xyz".
func New(baseURL, apiKey, pluginName string) (*Client, error) {
	u, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}
	if apiKey == "" {
		return nil, errors.New("apiKey is required")
	}
	if pluginName == "" {
		pluginName = "kavita-go"
	}
	return &Client{
		baseURL:    u,
		apiKey:     apiKey,
		pluginName: pluginName,
		http:       &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// SetHTTPClient lets callers swap the underlying http.Client (timeouts,
// transports for testing, etc.).
func (c *Client) SetHTTPClient(h *http.Client) { c.http = h }

// ---------------------------------------------------------------------------
// Auth
// ---------------------------------------------------------------------------

// Authenticate exchanges the API key for a JWT. Safe to call repeatedly; the
// returned token is cached on the client.
func (c *Client) Authenticate(ctx context.Context) error {
	q := url.Values{}
	q.Set("apiKey", c.apiKey)
	q.Set("pluginName", c.pluginName)

	var u UserDto
	if err := c.do(ctx, http.MethodPost, "/api/Plugin/authenticate?"+q.Encode(), nil, &u, false); err != nil {
		return fmt.Errorf("authenticate: %w", err)
	}
	c.mu.Lock()
	c.token = u.Token
	c.refreshToken = u.RefreshToken
	c.mu.Unlock()
	return nil
}

// RefreshToken swaps the cached refresh token for a fresh JWT. Most callers
// don't need this — do() retries on 401 by re-authenticating with the API key.
func (c *Client) RefreshToken(ctx context.Context) error {
	c.mu.RLock()
	body := TokenRequestDto{Token: c.token, RefreshToken: c.refreshToken}
	c.mu.RUnlock()

	var out TokenRequestDto
	if err := c.do(ctx, http.MethodPost, "/api/Account/refresh-token", body, &out, false); err != nil {
		return err
	}
	c.mu.Lock()
	c.token = out.Token
	c.refreshToken = out.RefreshToken
	c.mu.Unlock()
	return nil
}

// ---------------------------------------------------------------------------
// Request plumbing
// ---------------------------------------------------------------------------

// do performs an HTTP request, JSON-encoding body and JSON-decoding into out
// (either may be nil). When auth is true and the response is 401, it
// re-authenticates once and retries.
func (c *Client) do(ctx context.Context, method, path string, body, out any, auth bool) error {
	endpoint := c.baseURL.String() + path

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if auth {
		c.mu.RLock()
		tok := c.token
		c.mu.RUnlock()
		if tok == "" {
			return errors.New("not authenticated: call Authenticate first")
		}
		req.Header.Set("Authorization", "Bearer "+tok)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// One-shot re-auth on 401.
	if auth && resp.StatusCode == http.StatusUnauthorized {
		if err := c.Authenticate(ctx); err != nil {
			return fmt.Errorf("reauth after 401: %w", err)
		}
		return c.do(ctx, method, path, body, out, auth)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return &APIError{
			Status: resp.StatusCode,
			Body:   strings.TrimSpace(string(msg)),
			Path:   path,
		}
	}

	if out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// APIError is returned for non-2xx responses.
type APIError struct {
	Status int
	Body   string
	Path   string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("kavita: %s -> %d: %s", e.Path, e.Status, e.Body)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// APIKey returns the configured API key (handy for building image URLs).
func (c *Client) APIKey() string { return c.apiKey }

// BaseURL returns the configured server URL as a string.
func (c *Client) BaseURL() string { return c.baseURL.String() }

// SeriesCoverURL builds the unauthenticated cover-image URL for a series.
// Use this for `thumbnail` fields in your sidecar's JSON output — clients
// can fetch it directly.
func (c *Client) SeriesCoverURL(seriesID int) string {
	return fmt.Sprintf("%s/api/Image/series-cover?seriesId=%d&apiKey=%s",
		c.baseURL.String(), seriesID, url.QueryEscape(c.apiKey))
}

// ChapterCoverURL builds the chapter-level cover URL. For epub libraries
// this is the per-book cover.
func (c *Client) ChapterCoverURL(chapterID int) string {
	return fmt.Sprintf("%s/api/Image/chapter-cover?chapterId=%d&apiKey=%s",
		c.baseURL.String(), chapterID, url.QueryEscape(c.apiKey))
}
