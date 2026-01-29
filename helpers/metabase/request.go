package metabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// RequestOptions опции для запроса
type RequestOptions struct {
	Method  string
	Path    string
	Query   map[string]string
	Body    any
	Headers map[string]string
}

// Response ответ от API
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

func (c *Client) Request2(path string) ([]byte, error) {
	// Собираем URL
	fullURL, err := c.buildURL(path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	var bodyReader io.Reader
	req, err := http.NewRequest("POST", fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	fmt.Println(fullURL)
	req.Header.Set("X-API-Key", c.apiKey)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Читаем тело ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

func (c *Client) Request(options RequestOptions) (*Response, error) {
	// Собираем URL
	fullURL, err := c.buildURL(options.Path, options.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	// Подготавливаем тело запроса
	var bodyReader io.Reader
	if options.Body != nil {
		bodyBytes, err := json.Marshal(options.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Создаем HTTP запрос
	req, err := http.NewRequest(options.Method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Устанавливаем заголовки
	c.setHeaders(req, options.Headers)

	// Выполняем запрос
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Читаем тело ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Проверяем статус код
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
