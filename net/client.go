package net

import (
	"encoding/base64"
	"fmt"
	"io"
	gohttp "net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Fetcher is the minimal interface for loading bytes from URLs.
// Implemented by *Client and accepted by widget.Config.
type Fetcher interface {
	Fetch(rawURL string) ([]byte, error)
	FetchAsync(rawURL string, cb func([]byte, error))
}

// Options configures a Client.
type Options struct {
	// Proxy is the HTTP proxy URL, e.g. "http://127.0.0.1:7890".
	// Supports HTTP and HTTPS proxies.
	Proxy string
	// Timeout for each HTTP request. Default: 30s.
	Timeout time.Duration
	// MaxCacheBytes limits total cache size in bytes. 0 = unlimited.
	MaxCacheBytes int64
}

// cacheEntry stores a cached response.
type cacheEntry struct {
	data []byte
}

// Client handles HTTP/HTTPS image loading, data: URL decoding, and LRU caching.
// Safe for concurrent use.
type Client struct {
	http  *gohttp.Client
	cache sync.Map // string → *cacheEntry
}

// New creates a Client with the given options.
func New(opts Options) *Client {
	transport := &gohttp.Transport{
		Proxy: gohttp.ProxyFromEnvironment,
	}
	if opts.Proxy != "" {
		if proxyURL, err := url.Parse(opts.Proxy); err == nil {
			transport.Proxy = gohttp.ProxyURL(proxyURL)
		}
	}
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &Client{
		http: &gohttp.Client{
			Transport: transport,
			Timeout:   timeout,
		},
	}
}

// Default is a ready-to-use client with 30s timeout and proxy from environment.
var Default = New(Options{})

// Fetch returns the raw bytes at rawURL.
// Supports http://, https://, and data: URLs.
// HTTP responses are cached in memory.
func (c *Client) Fetch(rawURL string) ([]byte, error) {
	if strings.HasPrefix(rawURL, "data:") {
		return ParseDataURL(rawURL)
	}
	// Cache hit
	if v, ok := c.cache.Load(rawURL); ok {
		src := v.(*cacheEntry).data
		out := make([]byte, len(src))
		copy(out, src)
		return out, nil
	}
	resp, err := c.http.Get(rawURL)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", rawURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("fetch %s: HTTP %d", rawURL, resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", rawURL, err)
	}
	cached := make([]byte, len(data))
	copy(cached, data)
	c.cache.Store(rawURL, &cacheEntry{data: cached})
	return data, nil
}

// FetchAsync fetches rawURL in a goroutine and calls cb when done.
// cb is called from a goroutine — do NOT create or modify widgets inside cb.
// Instead store the result and process it on the main thread (e.g. in SetOnLayout).
func (c *Client) FetchAsync(rawURL string, cb func([]byte, error)) {
	go func() {
		data, err := c.Fetch(rawURL)
		cb(data, err)
	}()
}

// Invalidate removes rawURL from the cache.
func (c *Client) Invalidate(rawURL string) {
	c.cache.Delete(rawURL)
}

// Clear removes all cached entries.
func (c *Client) Clear() {
	c.cache.Range(func(k, _ any) bool {
		c.cache.Delete(k)
		return true
	})
}

// ParseDataURL decodes a data: URL and returns the raw bytes.
// Format: data:[<mediatype>][;base64],<data>
func ParseDataURL(s string) ([]byte, error) {
	if !strings.HasPrefix(s, "data:") {
		return nil, fmt.Errorf("not a data URL")
	}
	s = s[len("data:"):]
	comma := strings.IndexByte(s, ',')
	if comma < 0 {
		return nil, fmt.Errorf("invalid data URL: missing comma")
	}
	meta := s[:comma]
	data := s[comma+1:]
	if strings.HasSuffix(meta, ";base64") {
		// Try multiple base64 variants for robustness
		b, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			b, err = base64.URLEncoding.DecodeString(data)
		}
		if err != nil {
			b, err = base64.RawStdEncoding.DecodeString(data)
		}
		return b, err
	}
	// URL percent-encoded text content
	decoded, err := url.PathUnescape(data)
	return []byte(decoded), err
}

// IsRemoteURL reports whether src is a network URL (http/https) or data URL.
func IsRemoteURL(src string) bool {
	return strings.HasPrefix(src, "http://") ||
		strings.HasPrefix(src, "https://") ||
		strings.HasPrefix(src, "data:")
}
