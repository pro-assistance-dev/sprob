package search

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type SearchGroup struct {
	bun.BaseModel     `bun:"search_groups,alias:search_groups"`
	ID                uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	Key               string        `json:"key"`
	Label             string        `json:"label"`
	Order             int           `bun:"search_group_order" json:"order"`
	Active            bool          `bun:"-" json:"active"`
	Route             string        `json:"route"`
	Table             string        `bun:"search_group_table" json:"table"`
	SearchColumn      string        `json:"searchColumn"`
	LabelColumn       string        `json:"labelColumn"`
	ValueColumn       string        `json:"valueColumn"`
	DescriptionColumn string        `json:"descriptionColumn"`

	SearchElements         SearchElements         `bun:"-" json:"options"`
	SearchGroupMetaColumns SearchGroupMetaColumns `bun:"rel:has-many" json:"searchGroupMetaColumns"`
}

type SearchGroups []*SearchGroup

func (item *SearchGroup) BuildRoutes() {
	for i := range item.SearchElements {
		item.SearchElements[i].Value = fmt.Sprintf("%s/%s", item.Route, item.SearchElements[i].Value)
	}
}

func (item *SearchGroup) ParseMap(re map[string]interface{}) {
	for _, hit := range re["hits"].(map[string]interface{})["hits"].([]interface{}) {
		//index := hit.(map[string]interface{})["_index"]
		searchElement := SearchElement{}
		searchElement.Value = hit.(map[string]interface{})["_id"].(string)
		searchElement.Label = hit.(map[string]interface{})["_source"].(map[string]interface{})["name"].(string)
		searchElement.Description = hit.(map[string]interface{})["_source"].(map[string]interface{})["name"].(string)
		item.SearchElements = append(item.SearchElements, &searchElement)
	}
}

func (items SearchGroups) ParseMap(re map[string]interface{}) {
	for _, hit := range re["hits"].(map[string]interface{})["hits"].([]interface{}) {
		searchElement := SearchElement{}
		searchElement.Value = hit.(map[string]interface{})["_id"].(string)
		searchElement.Label = hit.(map[string]interface{})["_source"].(map[string]interface{})["name"].(string)
		searchElement.Description = hit.(map[string]interface{})["_source"].(map[string]interface{})["name"].(string)
		//item.SearchElements = append(item.SearchElements, &searchElement)
	}
}
