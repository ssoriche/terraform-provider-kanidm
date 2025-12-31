package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	// ErrNotFound indicates the requested resource was not found
	ErrNotFound = errors.New("resource not found")
	// ErrUnauthorized indicates authentication failed
	ErrUnauthorized = errors.New("unauthorized")
	// ErrForbidden indicates insufficient permissions
	ErrForbidden = errors.New("forbidden")
)

// Client provides methods to interact with the Kanidm API
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// ClientOption configures the Client
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithTimeout sets the HTTP client timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// NewClient creates a new Kanidm API client
func NewClient(baseURL, token string, opts ...ClientOption) *Client {
	c := &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// doRequest executes an HTTP request with proper error handling
func (c *Client) doRequest(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}

	if err := c.checkResponse(resp); err != nil {
		_ = resp.Body.Close()
		return nil, err
	}

	return resp, nil
}

// checkResponse validates HTTP response and returns typed errors
func (c *Client) checkResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrForbidden
	default:
		if len(body) > 0 {
			return fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, body)
		}
		return fmt.Errorf("API error (HTTP %d)", resp.StatusCode)
	}
}

// decodeResponse unmarshals the response body into the target
func decodeResponse(resp *http.Response, target any) error {
	defer func() { _ = resp.Body.Close() }()

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

// Entry represents a Kanidm resource with attributes
type Entry struct {
	Attrs map[string]any `json:"attrs"`
}

// GetString retrieves a string attribute, handling Kanidm's array-based attributes
func (e *Entry) GetString(key string) string {
	val, ok := e.Attrs[key]
	if !ok {
		return ""
	}

	switch v := val.(type) {
	case string:
		return v
	case []any:
		if len(v) > 0 {
			if s, ok := v[0].(string); ok {
				return s
			}
		}
	}

	return ""
}

// GetStringSlice retrieves a string slice attribute
func (e *Entry) GetStringSlice(key string) []string {
	val, ok := e.Attrs[key]
	if !ok {
		return nil
	}

	switch v := val.(type) {
	case []any:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	case []string:
		return v
	case string:
		return []string{v}
	}

	return nil
}

// CreateRequest represents a resource creation request
type CreateRequest struct {
	Attrs map[string]any `json:"attrs"`
}

// NewCreateRequest creates a new CreateRequest with the given attributes
func NewCreateRequest(attrs map[string]any) *CreateRequest {
	return &CreateRequest{Attrs: attrs}
}

// UpdateRequest represents a resource update request
type UpdateRequest struct {
	Attrs map[string]any `json:"attrs"`
}

// NewUpdateRequest creates a new UpdateRequest with the given attributes
func NewUpdateRequest(attrs map[string]any) *UpdateRequest {
	return &UpdateRequest{Attrs: attrs}
}
