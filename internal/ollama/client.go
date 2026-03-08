// Package ollama provides integration with the Ollama API for AI code generation.
package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents an Ollama API client
type Client struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

// Request represents a request to the Ollama API
type Request struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// Response represents a response from the Ollama API
type Response struct {
	Model     string `json:"model"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
	CreatedAt string `json:"created_at,omitempty"`
}

// ClientOption is a function that configures a Client
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

// NewClient creates a new Ollama API client
func NewClient(baseURL, model string, opts ...ClientOption) *Client {
	client := &Client{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// Generate generates a response from the Ollama API
func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	reqBody := Request{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", &APIError{Op: "marshal", Err: err}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", &APIError{Op: "create request", Err: err}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", &APIError{Op: "send request", Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", &APIError{
			Op:         "API request",
			StatusCode: resp.StatusCode,
			Err:        fmt.Errorf("Ollama returned status %d: %s", resp.StatusCode, string(body)),
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &APIError{Op: "read response", Err: err}
	}

	var ollamaResp Response
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", &APIError{Op: "parse response", Err: err}
	}

	return ollamaResp.Response, nil
}

// GenerateWithTimeout generates a response with a timeout
func (c *Client) GenerateWithTimeout(prompt string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return c.Generate(ctx, prompt)
}

// APIError represents an Ollama API error
type APIError struct {
	Op         string
	StatusCode int
	Err        error
}

func (e *APIError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("ollama %s (status %d): %v", e.Op, e.StatusCode, e.Err)
	}
	return fmt.Sprintf("ollama %s: %v", e.Op, e.Err)
}

func (e *APIError) Unwrap() error {
	return e.Err
}

// IsNotFound returns true if the error indicates the model was not found
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsConnectionError returns true if the error indicates a connection problem
func (e *APIError) IsConnectionError() bool {
	if e.Err == nil {
		return false
	}
	// Check for common connection errors
	errStr := e.Error()
	return contains(errStr, "connection refused") ||
		contains(errStr, "no such host") ||
		contains(errStr, "timeout")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
