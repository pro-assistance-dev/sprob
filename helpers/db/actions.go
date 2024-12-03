package db

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/iancoleman/strcase"
	"github.com/uptrace/bun/migrate"
)

func validateMigrateName(name string) error {
	if name == "" {
		return errors.New("migration name cannot be empty")
	}
	return nil
}

func createMigrationSQL(migrator *migrate.Migrator, name string) error {
	err := validateMigrateName(name)
	if err != nil {
		return err
	}
	nameInSnake := strcase.ToSnake(name)
	mf, err := migrator.CreateSQLMigrations(context.TODO(), nameInSnake)
	if err != nil {
		return err
	}
	fmt.Printf("created migration %s (%s)\n", mf[0].Name, mf[0].Path)
	return nil
}

func dropDatabase(migrator *migrate.Migrator) {
	_, err := migrator.DB().Exec(
		`DO $$ DECLARE
    r RECORD;
BEGIN
    FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = current_schema()) LOOP
        EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE';
    END LOOP;
END $$;`)
	if err != nil {
		log.Fatalln(err)
	}
}

func initMigration(migrator *migrate.Migrator) {
	err := migrator.Init(context.TODO())
	if err != nil {
		fmt.Println(err)
	}
	_, err = migrator.DB().Exec("create sequence bun_migration_locks_id_seq;")
	if err != nil {
		fmt.Println(err)
	}
	_, err = migrator.DB().Exec("create sequence bun_migrations_id_seq;")
	if err != nil {
		fmt.Println(err)
	}
	_, err = migrator.DB().Exec("alter table bun_migration_locks alter column id set default nextval('public.bun_migration_locks_id_seq');")
	if err != nil {
		fmt.Println(err)
	}
	_, err = migrator.DB().Exec("alter table bun_migrations alter column id set default nextval('public.bun_migrations_id_seq');")
	if err != nil {
		fmt.Println(err)
	}
	_, err = migrator.DB().Exec("alter sequence bun_migration_locks_id_seq owned by bun_migration_locks.id;")
	if err != nil {
		fmt.Println(err)
	}
	_, err = migrator.DB().Exec("alter sequence bun_migrations_id_seq owned by bun_migrations.id;")
	if err != nil {
		fmt.Println(err)
	}

	if err != nil {
		fmt.Println(err)
	}
}

func runMigration(migrator *migrate.Migrator) {
	group, err := migrator.Migrate(context.TODO())
	if err != nil {
		log.Fatalf("fail migrate: %s", err)
	}

	if group == nil || group.ID == 0 {
		fmt.Printf("there are no new migrations to run\n")
		return
	}

	fmt.Printf("migrated to %s\n", group)
}
