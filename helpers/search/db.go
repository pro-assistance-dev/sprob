package search

import (
	"context"
	"log"

	"github.com/pro-assistance-dev/sprob/models"

	"github.com/uptrace/bun"
)

func InitSearchGroupsTables(db *bun.DB) {
	_, err := db.NewCreateTable().Model((*models.SearchGroup)(nil)).IfNotExists().Exec(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	_, err = db.NewCreateTable().Model((*models.SearchGroupMetaColumn)(nil)).IfNotExists().Exec(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
}
