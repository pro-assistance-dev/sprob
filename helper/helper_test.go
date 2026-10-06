package helper

import (
	"context"
	dbsql "database/sql"
	"testing"

	coreMigrations "github.com/pro-assistance-dev/sprob/migrations"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/migrate"

	_ "modernc.org/sqlite"
)

// updateDB обязан ЗАВОДИТЬ таблицу bun_migrations сам: bun.Migrate() её читает,
// но не создаёт — на чистой БД это давало 42P01 и restart-loop сервиса.
func TestUpdateDB_CreatesMigrationsTableOnFreshDB(t *testing.T) {
	sqldb, err := dbsql.Open("sqlite", "file:updateDB-test?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqldb.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqldb.Close() })

	db := bun.NewDB(sqldb, sqlitedialect.New())
	migrator := migrate.NewMigrator(db, coreMigrations.Init())

	updateDB(migrator)

	var name string
	err = db.NewRaw("select name from sqlite_master where type='table' and name='bun_migrations'").
		Scan(context.Background(), &name)
	if err != nil {
		t.Fatalf("bun_migrations не создана: %v", err)
	}
	if name != "bun_migrations" {
		t.Fatalf("ожидалась bun_migrations, получено %q", name)
	}
}
