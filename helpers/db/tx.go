package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/uptrace/bun"
)

type txKey struct{}

func injectTx(ctx context.Context, tx bun.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func extractTx(ctx context.Context) *bun.Tx {
	if tx, ok := ctx.Value(txKey{}).(*bun.Tx); ok {
		return tx
	}
	return nil
}

func (i *DB) IDB(ctx context.Context) bun.IDB {
	tx := extractTx(ctx)
	if tx != nil {
		return tx
	}
	return i.DB
}

func (i *DB) WithinTransaction(ctx context.Context, tFunc func(ctx context.Context) error) error {
	tx, err := i.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	err = tFunc(injectTx(ctx, tx))
	if err != nil {
		if errRollback := tx.Rollback(); errRollback != nil {
			log.Printf("rollback transaction: %v", errRollback)
		}
		return err
	}

	if errCommit := tx.Commit(); errCommit != nil {
		log.Printf("commit transaction: %v", errCommit)
	}
	return nil
}
