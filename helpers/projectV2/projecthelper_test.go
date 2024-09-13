package project

import (
	"log"
	"testing"

	"github.com/pro-assistance/pro-assister/config"
)

func TestInitSchemas(t *testing.T) {
	conf, err := config.LoadTestConfig()
	if err != nil {
		log.Fatal(err)
	}
	p := NewProject(conf)
	t.Run("run", func(t *testing.T) {
		p.InitSchemas()
	})
}
