package metabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/pro-assistance-dev/sprob/config"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	SiteURL    string
	SecretKey  string
}

type RequestOptions struct {
	Method  string
	Path    string
	Query   map[string]string
	Body    any
	Headers map[string]string
}

type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

func NewClient(c config.Metabase) *Client {
	return &Client{
		baseURL:   c.URL,
		apiKey:    c.APIKey,
		SiteURL:   c.SiteURL,
		SecretKey: c.SecretKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) Get(path string, query map[string]string, headers map[string]string) (*Response, error) {
	return c.Request(RequestOptions{
		Method:  http.MethodGet,
		Path:    path,
		Query:   query,
		Headers: headers,
	})
}

func (c *Client) Post(path string, body any, headers map[string]string) (*Response, error) {
	return c.Request(RequestOptions{
		Method:  http.MethodPost,
		Path:    path,
		Body:    body,
		Headers: headers,
	})
}

func (c *Client) Request(options RequestOptions) (*Response, error) {
	fullURL, err := c.buildURL(options.Path, options.Query)
	if err != nil {
		return nil, fmt.Errorf("build URL failed: %w", err)
	}

	var bodyReader io.Reader
	if options.Body != nil {
		bodyBytes, err := json.Marshal(options.Body)
		if err != nil {
			return nil, fmt.Errorf("marshal body failed: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(options.Method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	c.setHeaders(req, options.Headers)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		return &Response{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			Body:       body,
		}, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       body,
	}, nil
}

func (c *Client) buildURL(path string, query map[string]string) (string, error) {
	fullURL, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		return "", err
	}

	if len(query) > 0 {
		params := url.Values{}
		for key, value := range query {
			params.Add(key, value)
		}
		fullURL = fullURL + "?" + params.Encode()
	}

	return fullURL, nil
}

func (c *Client) setHeaders(req *http.Request, additionalHeaders map[string]string) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	for key, value := range additionalHeaders {
		req.Header.Set(key, value)
	}
}

func (r *Response) ParseJSON(target any) error {
	if len(r.Body) == 0 {
		return fmt.Errorf("response body is empty")
	}
	return json.Unmarshal(r.Body, target)
}

func (r *Response) GetString() string {
	return string(r.Body)
}
