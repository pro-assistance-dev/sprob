package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"

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
	dsn := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=disable", i.config.DB, i.config.User, i.config.Password, i.config.Host, i.config.Port, i.config.Name)
	conn := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(conn, pgdialect.New())
	_, _ = db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	_, _ = db.Exec(`CREATE EXTENSION IF NOT EXISTS tablefunc;`)
	i.DB = db
}

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

func (i *DB) Dump() error {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("err")
	}
	exPath := filepath.Dir(filename)

	cmd := exec.Command("/bin/bash", filepath.Join(exPath, "dump_pg.sh"), i.config.Name, i.config.User, i.config.Password, i.config.RemoteUser, i.config.RemotePassword) //nolint:gosec
	return cmd.Run()
}
