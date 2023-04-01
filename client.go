package gpt35

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const ModelGpt35Turbo = "gpt-3.5-turbo"

const MaxTokensGpt35Turbo = 4096

const (
	RoleUser      RoleType = "user"
	RoleAssistant RoleType = "assistant"
	RoleSystem    RoleType = "system"
)

type Client struct {
	client *http.Client
	apiKey string
	url    string
}

// ClientOptions contains parameters for Client initialization
type ClientOptions struct {
	// HTTP transport
	Transport http.RoundTripper
	// Base url of OpenAI API
	URL string
}

type OptionFunc func(opts *ClientOptions) error

// WithTransport allows to override default client HTTP transport
func WithTransport(transport http.RoundTripper) OptionFunc {
	return func(opts *ClientOptions) error {
		if transport == nil {
			return errors.New("cannot set nil as HTTP transport")
		}

		opts.Transport = transport
		return nil
	}
}

// WithURL allows to override base url of OpenAPI
func WithURL(baseURL string) OptionFunc {
	return func(opts *ClientOptions) error {
		opts.URL = baseURL
		return nil
	}
}

func defaultOptions() *ClientOptions {
	return &ClientOptions{
		Transport: http.DefaultTransport,
		URL:       "https://api.openai.com/v1/chat/completions",
	}
}

// NewClient creates new OpenAI client
func NewClient(apiKey string, opts ...OptionFunc) (*Client, error) {
	params := defaultOptions()
	for _, opt := range opts {
		if err := opt(params); err != nil {
			return nil, fmt.Errorf("apply option: %w", err)
		}
	}

	return &Client{
		client: &http.Client{Transport: params.Transport},
		apiKey: apiKey,
		url:    params.URL,
	}, nil
}

// GetChat returns chat.
func (c *Client) GetChat(r *Request) (*Response, error) {
	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(r); err != nil {
		return nil, fmt.Errorf("encode json: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.url, buf)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http do: %w", err)
	}

	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("json decode: %w", err)
	}

	return &result, nil
}
