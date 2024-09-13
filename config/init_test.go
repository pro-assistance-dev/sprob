package config

import (
	"log"
	"testing"
)

func TestLoadTestConfig(t *testing.T) {
	t.Run("LoadTestConfig", func(t *testing.T) {
		_, err := LoadTestConfig()
		if err != nil {
			log.Fatal(err)
		}
	})
}
