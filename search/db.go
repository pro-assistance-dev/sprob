package search

import (
	"context"
	"github.com/uptrace/bun"
	"log"
)

func InitSearchGroupsTables(db *bun.DB) {
	_, err := db.NewCreateTable().Model((*SearchGroup)(nil)).IfNotExists().Exec(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	_, err = db.NewCreateTable().Model((*SearchGroupMetaColumn)(nil)).IfNotExists().Exec(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
}
