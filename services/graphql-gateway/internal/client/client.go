package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"maps"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	XRequestID = "X-Request-Id"
	XClientIP  = "X-Client-IP"
)

// Client is the main HTTP client for service-to-service communication
type Client struct {
	baseURL    string
	httpClient *http.Client
	headers    map[string]string
	logger     *log.Logger
	config     *ClientConfig
}

// ClientConfig holds client configuration
type ClientConfig struct {
	BaseURL       string
	Timeout       time.Duration
	MaxRetries    int
	RetryWaitTime time.Duration
	Headers       map[string]string
	Logger        *log.Logger
}

// NewClient creates a new HTTP client
func NewClient(config *ClientConfig) *Client {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryWaitTime == 0 {
		config.RetryWaitTime = 100 * time.Millisecond
	}
	if config.Logger == nil {
		logger := slog.NewLogLogger(slog.NewJSONHandler(os.Stdout, nil), slog.LevelDebug)
		config.Logger = logger
	}

	return &Client{
		baseURL: config.BaseURL,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		headers: config.Headers,
		logger:  config.Logger,
	}
}

// RequestOptions for making HTTP requests
type RequestOptions struct {
	Method  string
	Path    string
	Body    interface{}
	Headers map[string]string
	Query   map[string]string
	Context context.Context
}

// Response wrapper
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// Do makes an HTTP request
func (c *Client) Do(opts *RequestOptions) (*Response, error) {
	// Build URL
	url := c.baseURL + opts.Path

	// Add query params
	if len(opts.Query) > 0 {
		url += "?" + buildQueryString(opts.Query)
	}

	// Prepare body
	var bodyReader io.Reader
	if opts.Body != nil {
		jsonBody, err := json.Marshal(opts.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	// Create request
	req, err := http.NewRequestWithContext(opts.Context, opts.Method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Add client headers
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	// Add request-specific headers
	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	// Add trace ID if not present
	if req.Header.Get(XRequestID) == "" {
		req.Header.Set(XRequestID, generateRequestID())
	}

	// Log request
	c.logger.Print("HTTP Request",
		"method", opts.Method,
		"url", url,
		"headers", req.Header,
	)

	// Execute with retries
	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		resp, err = c.httpClient.Do(req)
		if err == nil {
			break
		}

		lastErr = err
		time.Sleep(c.config.RetryWaitTime)
	}

	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", c.config.MaxRetries, lastErr)
	}
	defer resp.Body.Close()

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Log response
	c.logger.Print("HTTP Response",
		"status", resp.StatusCode,
		"body", string(body),
	)

	// Check status code
	if resp.StatusCode >= 400 {
		return &Response{
				StatusCode: resp.StatusCode,
				Headers:    resp.Header,
				Body:       body,
			}, &HTTPError{
				StatusCode: resp.StatusCode,
				Body:       body,
				Message:    fmt.Sprintf("HTTP %d", resp.StatusCode),
			}
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       body,
	}, nil
}

// Convenience methods
func (c *Client) Get(ctx context.Context, path string, options ...RequestOption) (*Response, error) {
	opts := &RequestOptions{
		Method:  http.MethodGet,
		Path:    path,
		Context: ctx,
	}
	for _, opt := range options {
		opt(opts)
	}
	return c.Do(opts)
}

func (c *Client) Post(ctx context.Context, path string, body interface{}, options ...RequestOption) (*Response, error) {
	opts := &RequestOptions{
		Method:  http.MethodPost,
		Path:    path,
		Body:    body,
		Context: ctx,
	}
	for _, opt := range options {
		opt(opts)
	}
	return c.Do(opts)
}

func (c *Client) Put(ctx context.Context, path string, body interface{}, options ...RequestOption) (*Response, error) {
	opts := &RequestOptions{
		Method:  http.MethodPut,
		Path:    path,
		Body:    body,
		Context: ctx,
	}
	for _, opt := range options {
		opt(opts)
	}
	return c.Do(opts)
}

func (c *Client) Delete(ctx context.Context, path string, options ...RequestOption) (*Response, error) {
	opts := &RequestOptions{
		Method:  http.MethodDelete,
		Path:    path,
		Context: ctx,
	}
	for _, opt := range options {
		opt(opts)
	}
	return c.Do(opts)
}

// RequestOption functional option
type RequestOption func(*RequestOptions)

func WithHeaders(headers map[string]string) RequestOption {
	return func(opts *RequestOptions) {
		if opts.Headers == nil {
			opts.Headers = make(map[string]string)
		}
		maps.Copy(opts.Headers, headers)
	}
}

func WithQuery(query map[string]string) RequestOption {
	return func(opts *RequestOptions) {
		opts.Query = query
	}
}

// HTTPError represents an HTTP error
type HTTPError struct {
	StatusCode int
	Body       []byte
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// Helper functions
func buildQueryString(params map[string]string) string {
	if len(params) == 0 {
		return ""
	}
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	return values.Encode()
}

func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
