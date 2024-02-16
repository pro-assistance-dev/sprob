package models

import "fmt"

type SearchModel struct {
	Suggester       bool           `json:"suggester"`
	SearchElements  SearchElements `bun:"-" json:"options"`
	Query           string         `json:"query"`
	MustBeTranslate bool           `json:"mustBeTranslate"`
	TranslitQuery   string         `json:"translitQuery"`
	SearchGroupID   string         `json:"searchGroupId"`
	SearchGroups    SearchGroups   `json:"searchGroups"`
	SearchGroup     *SearchGroup   `json:"searchGroup"`
}

func (item *SearchModel) findGroup(groupTable string) SearchGroup {
	group := SearchGroup{}
	for _, searchGroup := range item.SearchGroups {
		if searchGroup.Table == groupTable {
			group = *searchGroup
			break
		}
	}
	return group
}

func (item *SearchModel) createSearchElement(resultElement interface{}, group SearchGroup) *SearchElement {
	searchElement := SearchElement{}
	searchElement.Value = resultElement.(map[string]interface{})["_id"].(string)
	searchElement.Label = resultElement.(map[string]interface{})["_source"].(map[string]interface{})[group.LabelColumn].(string)
	if !item.Suggester && group.DescriptionColumn != "" && resultElement.(map[string]interface{})["_source"].(map[string]interface{})[group.DescriptionColumn] != nil {
		searchElement.Description = resultElement.(map[string]interface{})["_source"].(map[string]interface{})[group.DescriptionColumn].(string)
	}
	searchElement.Route = fmt.Sprintf("%s/%s", group.Route, searchElement.Value)
	if len(group.SearchGroupMetaColumns) > 0 {
		for _, metaCol := range group.SearchGroupMetaColumns {
			if resultElement.(map[string]interface{})["_source"].(map[string]interface{})[metaCol.Name] != nil {
				searchElementMeta := SearchElementMeta{}
				searchElementMeta.Value = resultElement.(map[string]interface{})["_source"].(map[string]interface{})[metaCol.Name].(string)
				searchElementMeta.Name = metaCol.Name
				searchElement.SearchElementMetas = append(searchElement.SearchElementMetas, &searchElementMeta)
			}
		}
	}

	searchElement.SearchGroup = &group
	return &searchElement
}

func (item *SearchModel) parseSuggest(re map[string]interface{}) {
	for _, hit := range re["hits"].(map[string]interface{})["hits"].([]interface{}) {
		index := hit.(map[string]interface{})["_index"]
		group := item.findGroup(index.(string))
		if group.Label != "" {
			searchElement := item.createSearchElement(hit, group)
			item.SearchElements = append(item.SearchElements, searchElement)
		}
	}
}

func (item *SearchModel) ParseMap(re map[string]interface{}) {
	if item.Suggester {
		item.parseSuggest(re)
		return
	}
	for _, hit := range re["hits"].(map[string]interface{})["hits"].([]interface{}) {
		index := hit.(map[string]interface{})["_index"]
		group := item.findGroup(index.(string))
		searchElement := item.createSearchElement(hit, group)
		item.SearchGroup.SearchElements = append(item.SearchGroup.SearchElements, searchElement)
	}
}

func (item *SearchModel) BuildQuery() (map[string]interface{}, []string) {
	fields := make([]string, 0)
	indexes := make([]string, 0)
	for _, group := range item.SearchGroups {
		if group.Active && group.Label != "" {
			fields = append(fields, group.SearchColumn)
			indexes = append(indexes, group.Table)
		}
	}
	fields = append(fields, "*")
	matchStatement := map[string]interface{}{
		"multi_match": map[string]interface{}{
			"query":          item.Query,
			"fields":         fields,
			"analyzer":       "rus_anal",
			"fuzziness":      1000,
			"max_expansions": 1000,
		},
	}
	query := map[string]interface{}{
		"query": matchStatement,
	}
	return query, indexes
}
