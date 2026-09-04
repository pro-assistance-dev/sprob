package migrations

import (
	"fmt"

	"github.com/uptrace/bun/migrate"
)

func Init() *migrate.Migrations {
	migrations := migrate.NewMigrations()
	if err := migrations.DiscoverCaller(); err != nil {
		fmt.Println(err)
	}
	return migrations
}
