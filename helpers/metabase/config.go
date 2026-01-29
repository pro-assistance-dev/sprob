package metabase

import "time"

type ClientConfig struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
	Headers map[string]string
}
