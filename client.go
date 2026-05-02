// Package postio is the Go SDK for the Postio API — UK address, email,
// and phone validation, backed by Royal Mail PAF and Ordnance Survey.
//
// Quick start:
//
//	client, err := postio.NewClient(postio.WithAPIKey("pk_live_..."))
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	result, err := client.Address.Search(context.Background(), "downing street", nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, hit := range result.Results {
//	    fmt.Println(hit.UDPRN, hit.Suggestion)
//	}
//
// The API key may also be supplied via the POSTIO_API_KEY environment
// variable, in which case WithAPIKey is optional.
package postio

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// DefaultBaseURL is the production Postio API endpoint.
const DefaultBaseURL = "https://api.postio.co.uk/v1"

// DefaultTimeout is the per-request HTTP timeout when no explicit
// http.Client is provided.
const DefaultTimeout = 10 * time.Second

const (
	defaultMaxRetries = 2
	defaultBaseDelay  = 500 * time.Millisecond
	defaultCapDelay   = 8 * time.Second
)

// Client is the top-level Postio API client. Safe for concurrent use.
type Client struct {
	apiKey       string
	baseURL      string
	httpClient   *http.Client
	extraHeaders map[string]string
	maxRetries   int
	baseDelay    time.Duration
	capDelay     time.Duration

	// Resource namespaces.
	Address *AddressService
	Email   *EmailService
	Phone   *PhoneService
}

// Option configures the Client at construction time.
type Option func(*Client) error

// WithAPIKey sets the API key explicitly. Overrides POSTIO_API_KEY env var.
func WithAPIKey(key string) Option {
	return func(c *Client) error {
		c.apiKey = key
		return nil
	}
}

// WithBaseURL overrides the base URL. Default is DefaultBaseURL.
// Useful for local testing or stage-api.postio.co.uk.
func WithBaseURL(url string) Option {
	return func(c *Client) error {
		c.baseURL = strings.TrimRight(url, "/")
		return nil
	}
}

// WithHTTPClient supplies a custom *http.Client (for proxies, custom
// transports, etc.). The client's Timeout is honoured as-is; pass a
// configured client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) error {
		if hc == nil {
			return errors.New("postio: WithHTTPClient: client cannot be nil")
		}
		c.httpClient = hc
		return nil
	}
}

// WithTimeout sets the per-request timeout when using the default
// http.Client. Ignored if WithHTTPClient is also passed.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) error {
		c.httpClient.Timeout = d
		return nil
	}
}

// WithRetries configures the retry policy. Pass 0 to disable retries.
// Default: 2 retries, exp backoff 500ms → 8s with full jitter.
func WithRetries(max int) Option {
	return func(c *Client) error {
		if max < 0 {
			return errors.New("postio: WithRetries: count must be >= 0")
		}
		c.maxRetries = max
		return nil
	}
}

// WithRetryBackoff overrides the exponential backoff parameters.
func WithRetryBackoff(base, cap time.Duration) Option {
	return func(c *Client) error {
		c.baseDelay = base
		c.capDelay = cap
		return nil
	}
}

// WithHeader adds a header to every request. x-api-key and accept
// cannot be overridden.
func WithHeader(key, value string) Option {
	return func(c *Client) error {
		if c.extraHeaders == nil {
			c.extraHeaders = make(map[string]string)
		}
		c.extraHeaders[key] = value
		return nil
	}
}

// NewClient constructs a Postio client. If WithAPIKey is not passed,
// the POSTIO_API_KEY environment variable is used.
func NewClient(opts ...Option) (*Client, error) {
	c := &Client{
		baseURL:    DefaultBaseURL,
		httpClient: &http.Client{Timeout: DefaultTimeout},
		maxRetries: defaultMaxRetries,
		baseDelay:  defaultBaseDelay,
		capDelay:   defaultCapDelay,
	}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	if c.apiKey == "" {
		c.apiKey = os.Getenv("POSTIO_API_KEY")
	}
	if c.apiKey == "" {
		return nil, errors.New("postio: api key is required (use WithAPIKey or set POSTIO_API_KEY)")
	}
	c.Address = &AddressService{client: c}
	c.Email = &EmailService{client: c}
	c.Phone = &PhoneService{client: c}
	return c, nil
}

// Connect calls /connect — a free health probe that confirms the API
// is reachable and the key is valid.
func (c *Client) Connect(ctx context.Context) (*ConnectSuccess, error) {
	var out ConnectSuccess
	if err := c.do(ctx, "/connect", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) do(ctx context.Context, path string, query url.Values, out any) error {
	endpoint := c.baseURL + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		err := c.attempt(ctx, endpoint, out)
		if err == nil {
			return nil
		}

		var apiErr *Error
		if errors.As(err, &apiErr) && shouldRetry(apiErr) && attempt < c.maxRetries {
			lastErr = err
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(jitterBackoff(c.baseDelay, c.capDelay, attempt)):
			}
			continue
		}
		return err
	}
	return lastErr
}

func (c *Client) attempt(ctx context.Context, endpoint string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return &Error{Code: "request_build_error", Message: err.Error(), Cause: err}
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("accept", "application/json")
	req.Header.Set("user-agent", fmt.Sprintf("postio-go/%s", Version))
	req.Header.Set("x-postio-client", fmt.Sprintf("postio-go/%s", Version))
	for k, v := range c.extraHeaders {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Distinguish timeout vs other transport errors.
		if errors.Is(err, context.DeadlineExceeded) || isTimeoutError(err) {
			return &Error{Code: "request_timeout", Message: "request timed out", Cause: err}
		}
		return &Error{Code: "network_error", Message: err.Error(), Cause: err}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &Error{Status: resp.StatusCode, Code: "read_error", Message: err.Error(), Cause: err}
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		preview := string(body)
		if len(preview) > 500 {
			preview = preview[:500]
		}
		return &Error{
			Status:  resp.StatusCode,
			Code:    "unexpected_content_type",
			Message: fmt.Sprintf("unexpected content-type %q", contentType),
			Details: preview,
		}
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if err := json.Unmarshal(body, out); err != nil {
			return &Error{Status: resp.StatusCode, Code: "parse_error", Message: err.Error(), Cause: err}
		}
		return nil
	}

	// Non-2xx: parse the envelope and surface a typed error.
	var env ErrorEnvelope
	_ = json.Unmarshal(body, &env)
	apiErr := &Error{
		Status:    resp.StatusCode,
		Code:      env.Error,
		Message:   env.Error,
		RequestID: env.Meta.RequestID,
		Envelope:  &env,
	}
	if env.Error == "" {
		apiErr.Message = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}
	if env.Details != nil {
		apiErr.Details = *env.Details
	}
	if resp.StatusCode == 429 {
		if h := resp.Header.Get("Retry-After"); h != "" {
			if seconds, err := strconv.ParseFloat(h, 64); err == nil {
				apiErr.RetryAfter = seconds
			}
		}
	}
	return apiErr
}

func shouldRetry(e *Error) bool {
	if e.Code == "request_timeout" || e.Code == "network_error" {
		return true
	}
	switch e.Status {
	case 408, 409, 429, 500, 502, 503, 504:
		return true
	}
	return false
}

func jitterBackoff(base, cap time.Duration, attempt int) time.Duration {
	exp := base << attempt
	if cap > 0 && exp > cap {
		exp = cap
	}
	if exp <= 0 {
		return 0
	}
	return time.Duration(rand.Int64N(int64(exp)))
}

// isTimeoutError reports whether err looks like a network timeout
// (covers net.Error.Timeout and the generic transport timeouts).
func isTimeoutError(err error) bool {
	type timeouter interface{ Timeout() bool }
	var t timeouter
	if errors.As(err, &t) {
		return t.Timeout()
	}
	return false
}
