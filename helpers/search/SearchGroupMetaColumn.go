package search

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type SearchGroupMetaColumn struct {
	bun.BaseModel `bun:"search_group_meta_columns,alias:search_group_meta_columns"`
	ID            uuid.UUID     `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	Label         string        `json:"label"`
	Name          string        `json:"name"`
	SearchGroup   *SearchGroup  `bun:"rel:belongs-to" json:"searchGroup"`
	SearchGroupID uuid.NullUUID `bun:"type:uuid" json:"searchGroupId"`
}

type SearchGroupMetaColumns []*SearchGroupMetaColumn
