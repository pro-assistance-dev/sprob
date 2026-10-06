package migrations

import (
	"fmt"

	"github.com/uptrace/bun/migrate"
)

// Init возвращает миграции модуля survey. Порядок относительно forms
// задаёт ПОТРЕБИТЕЛЬ: SQL survey ссылается на таблицы форм (forms/form_fills/...),
// поэтому forms.Init() обязан идти раньше (bun-migrate гоняет группы по порядку слайса).
func Init() *migrate.Migrations {
	migrations := migrate.NewMigrations()
	if err := migrations.DiscoverCaller(); err != nil {
		fmt.Println(err)
	}
	return migrations
}
