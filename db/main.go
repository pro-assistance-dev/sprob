package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/uptrace/bun/migrate"
	"log"

	"github.com/uptrace/bun/extra/bundebug"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/pgdriver"

	"github.com/pro-assistance/pro-assister/config"
)

type DB struct {
	config config.DB
	DB     *bun.DB
}

func NewDBHelper(config config.DB) *DB {
	return &DB{config: config}
}

func (i *DB) InitDB() *bun.DB {
	dsn := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=disable", i.config.DB, i.config.User, i.config.Password, i.config.Host, i.config.Port, i.config.Name)
	conn := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(conn, sqlitedialect.New())
	db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	_, _ = db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	_, _ = db.Exec(`CREATE EXTENSION IF NOT EXISTS tablefunc;`)
	return db
}

func (i *DB) DoAction(migrations *migrate.Migrations, name *string, action *string) {
	migrator := migrate.NewMigrator(i.DB, migrations)
	switch *action {
	case "init":
		initMigration(migrator)
	case "dropDatabase":
		dropDatabase(migrator)
	case "create":
		createMigrationSql(migrator, name)
	case "migrate":
		runMigration(migrator)
	case "status":
		ms, err := migrator.MigrationsWithStatus(context.TODO())
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("migrations: %s\n", ms)
		fmt.Printf("unapplied migrations: %s\n", ms.Unapplied())
		fmt.Printf("last migration group: %s\n", ms.LastGroup())
	default:
		log.Fatal("cannot parse action")
	}
}
