// Package summary turns a commit subject into a short sentence for the
// developer's close circle, using CLIProxy running on the same machine.
//
// Only the commit subject is sent, and only to a loopback address: never code,
// file names or anything else (see docs/summary.md and docs/privacy.md).
package summary

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Defaults from docs/summary.md.
const (
	DefaultURL     = "http://127.0.0.1:8317"
	DefaultModel   = "claude-haiku-5-5"
	DefaultTimeout = 8 * time.Second
)

const (
	anthropicVersion = "2023-06-01"
	maxTokens        = 120
	maxResponse      = 64 << 10
	cacheSize        = 256
)

const systemPrompt = `Bir yazılımcının commit mesajını, yazılımcının teknik olmayan yakın çevresinin (aile, arkadaş, partner) anlayacağı tek bir Türkçe cümleye çevir.
Kurallar:
- Tek cümle, en fazla 140 karakter.
- Teknik terim, dosya adı, kod veya ürün adı kullanma.
- Üçüncü tekil şahısla yaz (ör. "Uygulamanın giriş ekranındaki bir hatayı düzeltti.").
- Yalnızca cümleyi yaz; açıklama, tırnak veya emoji ekleme.`

// ErrNotLocal means the configured CLIProxy address is not on this machine.
var ErrNotLocal = errors.New("CLIProxy address must be on this machine (127.0.0.1, ::1 or localhost)")

// Options configures a Client.
type Options struct {
	URL     string
	APIKey  string
	Model   string
	Timeout time.Duration
}

// Client asks CLIProxy for commit summaries. It is safe for concurrent use.
type Client struct {
	endpoint string
	apiKey   string
	model    string
	http     *http.Client
	cache    *cache
}

// New returns a Client. It fails if opts.URL is not a loopback address.
func New(opts Options) (*Client, error) {
	if opts.URL == "" {
		opts.URL = DefaultURL
	}
	if opts.Model == "" {
		opts.Model = DefaultModel
	}
	if opts.Timeout <= 0 {
		opts.Timeout = DefaultTimeout
	}
	base, err := localURL(opts.URL)
	if err != nil {
		return nil, err
	}
	return &Client{
		endpoint: base.JoinPath("v1", "messages").String(),
		apiKey:   opts.APIKey,
		model:    opts.Model,
		http: &http.Client{
			Timeout: opts.Timeout,
			// A redirect could send the commit subject off the machine.
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
		cache: newCache(cacheSize),
	}, nil
}

// localURL parses raw and accepts it only if it points at this machine.
func localURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("parse CLIProxy address: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("CLIProxy address must start with http:// or https://, got %q", raw)
	}
	host := u.Hostname()
	if host == "localhost" {
		return u, nil
	}
	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		return u, nil
	}
	return nil, fmt.Errorf("%w: got %q", ErrNotLocal, host)
}

type request struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system"`
	Messages  []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type response struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

// Summarize returns the sentence for a commit message. Only the first line,
// at most MaxInput characters, is sent. The same subject is not sent twice.
func (c *Client) Summarize(ctx context.Context, commitMessage string) (string, error) {
	subject := Subject(commitMessage)
	if subject == "" {
		return "", errors.New("empty commit message")
	}
	if s, ok := c.cache.get(subject); ok {
		return s, nil
	}
	s, err := c.ask(ctx, subject)
	if err != nil {
		return "", err
	}
	c.cache.put(subject, s)
	return s, nil
}

func (c *Client) ask(ctx context.Context, subject string) (string, error) {
	body, err := json.Marshal(request{
		Model:     c.model,
		MaxTokens: maxTokens,
		System:    systemPrompt,
		Messages:  []message{{Role: "user", Content: subject}},
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("anthropic-version", anthropicVersion)
	if c.apiKey != "" {
		req.Header.Set("x-api-key", c.apiKey)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("reach CLIProxy: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxResponse))
	if err != nil {
		return "", fmt.Errorf("read CLIProxy answer: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("CLIProxy answered %s", resp.Status)
	}
	var r response
	if err := json.Unmarshal(data, &r); err != nil {
		return "", fmt.Errorf("parse CLIProxy answer: %w", err)
	}
	var text strings.Builder
	for _, part := range r.Content {
		if part.Type == "text" {
			text.WriteString(part.Text)
		}
	}
	s := clean(text.String())
	if s == "" {
		return "", errors.New("CLIProxy returned no text")
	}
	return s, nil
}
