package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/uptrace/bun/migrate"

	"github.com/uptrace/bun"

	"github.com/uptrace/bun/dialect/pgdialect"
	// "github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/pgdriver"

	"github.com/pro-assistance-dev/sprob/config"
)

type DB struct {
	config  config.DB
	DB      *bun.DB
	Verbose bool
}

func NewDB(config config.DB) *DB {
	verbose, err := strconv.ParseBool(config.Verbose)
	if err != nil {
		verbose = false
	}
	h := &DB{config: config, Verbose: verbose}
	h.initDB()
	return h
}

func (i *DB) initDB() {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", i.config.User, i.config.Password, i.config.Host, i.config.Port, i.config.Name)
	conn := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(conn, pgdialect.New())

	// ⚠️ Пул соединений. Без настроек `database/sql` держит соединения БЕЗ
	// ВРЕМЕННОГО ЛИМИТА и без продления: после рестарта БД (или её контейнера)
	// в пуле остаются «мёртвые» сокеты, и запрос падает
	// `read tcp …:5432: i/o timeout` → 500 у клиента (найдено 07.11: список
	// работников hr пустел после `docker restart hr-db`, `drv=…` в логе не было).
	// Продление по возрасту заставляет пул переподключаться сам.
	db.SetConnMaxLifetime(connMaxLifetime)
	db.SetConnMaxIdleTime(connMaxIdleTime)
	db.SetMaxIdleConns(connMaxIdle)
	db.SetMaxOpenConns(connMaxOpen)

	_, _ = db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	_, _ = db.Exec(`CREATE EXTENSION IF NOT EXISTS tablefunc;`)
	i.DB = db
}

const (
	// connMaxLifetime — возраст соединения, после которого оно закрывается
	// при возврате в пул (докер/k8s рвут «долгие» сокеты тише, чем мы узнаём).
	connMaxLifetime = 30 * time.Minute
	connMaxIdleTime = 5 * time.Minute
	connMaxIdle     = 10
	connMaxOpen     = 50
)

func (i *DB) DoAction(migrations []*migrate.Migrations, name string, action *string) error {
	if len(migrations) == 0 {
		return errors.New("no migrations modules")
	}
	migrator := migrate.NewMigrator(i.DB, migrations[0])
	var err error
	initMigration(migrator)
	switch *action {
	case "dropDatabase":
		dropDatabase(migrator)
	case "create":
		err = createMigrationSQL(migrator, name)
	case "migrate":
		for _, migration := range migrations {
			migrator = migrate.NewMigrator(i.DB, migration)
			runMigration(migrator)
		}
	case "status":
		ms, err := migrator.MigrationsWithStatus(context.TODO())
		if err != nil {
			return err
		}
		fmt.Printf("migrations: %s\n", ms)
		fmt.Printf("unapplied migrations: %s\n", ms.Unapplied())
		fmt.Printf("last migration group: %s\n", ms.LastGroup())
	default:
		return errors.New("cannot parse action")
	}
	return err
}
