package helper

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pro-assistance/pro-assister/config"
	"github.com/uptrace/bun/migrate"
)

func pusto(_ http.Handler, _ *Helper) {
	fmt.Println("pusto")
}
func Test_Run(t *testing.T) {

	conf, err := config.LoadTestConfig()
	if err != nil {
		t.Errorf("Ошибка: %s", err)
	}
	hl := NewHelper(*conf)
	m := migrate.NewMigrations()
	router := gin.New()
	router.Use(gin.Recovery())

	t.Run("run", func(t *testing.T) {
		res := hl.Run(m, router, pusto)
		if Run != res {
			t.Errorf("ожидается: %s, результат: %d", Run, res)
		}
	})
	t.Run("dump", func(t *testing.T) {
		res := hl.Run(m, router, pusto)
		if Dump != res {
			t.Errorf("ожидается: %s, результат: %d", Dump, res)
		}
	})
	t.Run("migration", func(t *testing.T) {
		res := hl.Run(m, router, pusto)
		if Migrate != res {
			t.Errorf("ожидается: %s, результат: %d", Migrate, res)
		}
	})
	t.Run("listen", func(t *testing.T) {
		res := hl.Run(m, router, pusto)
		if Listen != res {
			t.Errorf("ожидается: %s, результат: %d", Listen, res)
		}
	})
}
