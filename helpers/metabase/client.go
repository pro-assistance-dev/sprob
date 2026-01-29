package metabase

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pro-assistance-dev/sprob/config"
)

// const METABASE_API_KEY = "mb_dj8BXmYdz6tapw8fyIUYlRuZk66FdmyHaQmt1SZSA6U="

// Client представляет клиент для работы с Metabase API
type Client struct {
	baseURL    string
	apiKey     string
	dbID       string
	httpClient *http.Client
	headers    map[string]string
}

// NewClient создает новый клиент Metabase
func NewClient(config *config.Metabase) *Client {
	return &Client{
		baseURL: "http://metabase:3000",
		// apiKey:  config.APIKey,
		apiKey: config.APIKey,
		dbID:   config.DBID,
		httpClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:        100,
				IdleConnTimeout:     90 * time.Second,
				DisableCompression:  true,
				MaxIdleConnsPerHost: 100,
			},
		},
	}
}

// Get выполняет GET запрос
func (c *Client) Get(path string, query map[string]string, headers map[string]string) (*Response, error) {
	return c.Request(RequestOptions{
		Method:  http.MethodGet,
		Path:    path,
		Query:   query,
		Headers: headers,
	})
}

// Post выполняет POST запрос
func (c *Client) Post(path string, body any, headers map[string]string) (*Response, error) {
	return c.Request(RequestOptions{
		Method:  http.MethodPost,
		Path:    path,
		Body:    body,
		Headers: headers,
	})
}

// Put выполняет PUT запрос
func (c *Client) Put(path string, body any, headers map[string]string) (*Response, error) {
	return c.Request(RequestOptions{
		Method:  http.MethodPut,
		Path:    path,
		Body:    body,
		Headers: headers,
	})
}

// Delete выполняет DELETE запрос
func (c *Client) Delete(path string, headers map[string]string) (*Response, error) {
	return c.Request(RequestOptions{
		Method:  http.MethodDelete,
		Path:    path,
		Headers: headers,
	})
}

// ParseJSON парсит JSON из ответа в структуру
func (r *Response) ParseJSON(target any) error {
	if len(r.Body) == 0 {
		return fmt.Errorf("response body is empty")
	}
	return json.Unmarshal(r.Body, target)
}

// GetString возвращает тело ответа как строку
func (r *Response) GetString() string {
	return string(r.Body)
}

// buildURL строит полный URL с параметрами запроса
func (c *Client) buildURL(path string, query map[string]string) (string, error) {
	// Очищаем путь от слэша в начале
	path = strings.TrimPrefix(path, "/")

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

// setHeaders устанавливает заголовки запроса
func (c *Client) setHeaders(req *http.Request, additionalHeaders map[string]string) {
	// Базовые заголовки
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	// Пользовательские заголовки из конфига
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	for key, value := range additionalHeaders {
		req.Header.Set(key, value)
	}
}

// Helper методы для удобства

// ExecuteQuery выполняет SQL запрос к базе данных
func (c *Client) ExecuteQuery(sql string, parameters map[string]any) (*Response, error) {
	query := map[string]any{
		"query": sql,
	}

	if parameters != nil {
		query["parameters"] = parameters
	}

	body := map[string]any{
		"database": c.dbID,
		"type":     "native",
		"native":   query,
	}

	return c.Post("/api/card/query/native", body, nil)
}

// GetQuestionResults получает результаты сохраненного вопроса
func (c *Client) GetQuestionResults(questionID int, parameters map[string]any) (*Response, error) {
	path := fmt.Sprintf("/api/card/%d/query", questionID)
	body := map[string]any{
		"parameters": parameters,
	}

	return c.Post(path, body, nil)
}

//
// // ListDatabases получает список баз данных
// func (c *Client) ListDatabases() (*Response, error) {
// 	return c.Get("/api/database", nil, nil)
// }
//
// // GetDatabase получает информацию о базе данных
// func (c *Client) GetDatabase(databaseID int) (*Response, error) {
// 	path := fmt.Sprintf("/api/database/%d", databaseID)
// 	return c.Get(path, nil, nil)
// }
//
// // Search выполняет поиск в Metabase
// func (c *Client) Search(query string, models []string, limit int) (*Response, error) {
// 	params := map[string]string{
// 		"q": query,
// 	}
//
// 	if models != nil {
// 		modelsJSON, _ := json.Marshal(models)
// 		params["models"] = string(modelsJSON)
// 	}
//
// 	if limit > 0 {
// 		params["limit"] = fmt.Sprintf("%d", limit)
// 	}
//
// 	return c.Get("/api/search", params, nil)
// }
