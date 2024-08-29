package search

import (
	"context"
	"fmt"

	"github.com/pro-assistance/pro-assister/models"
)

func (r *Repository) GetGroupByKey(c context.Context, key string) (*models.SearchGroup, error) {
	item := models.SearchGroup{}
	query := r.helper.DB.IDB(c).NewSelect().Model(&item).
		Relation("SearchGroupMetaColumns").Where("key = ?", key)

	err := query.Scan(c)
	return &item, err
}

func (r *Repository) GetGroups(c context.Context, groupID string) (models.SearchGroups, error) {
	items := make(models.SearchGroups, 0)
	query := r.helper.DB.IDB(c).NewSelect().Model(&items).
		Relation("SearchGroupMetaColumns").
		Order("search_group_order").Where("route is not null")

	if groupID != "" {
		query = query.Where("id = ?", groupID)
	}
	err := query.Scan(c)
	return items, err
}

func (r *Repository) Search(c context.Context, searchModel *models.SearchModel) error {
	querySelect := fmt.Sprintf("SELECT %s.%s as value, substring(%s for 40) as label", searchModel.SearchGroup.Table, searchModel.SearchGroup.ValueColumn, searchModel.SearchGroup.LabelColumn)
	queryFrom := fmt.Sprintf("FROM %s", searchModel.SearchGroup.Table)
	join := ""

	condition := fmt.Sprintf("where replace(regexp_replace(%s, '[^а-яА-Яa-zA-Z0-9. ]', '', 'g'), ' ' , '') ILIKE %s", searchModel.SearchGroup.SearchColumn, "'%"+searchModel.Query+"%'")
	conditionTranslitToRu := fmt.Sprintf("or replace(regexp_replace(%s, '[^а-яА-Яa-zA-Z0-9. ]', '', 'g'), ' ', '') ILIKE %s", searchModel.SearchGroup.SearchColumn, "'%"+r.helper.Util.TranslitToRu(searchModel.Query)+"%'")
	conditionTranslitToEng := fmt.Sprintf("or replace(regexp_replace(%s, '[^а-яА-Яa-zA-Z0-9. ]', '', 'g'), ' ', '') ILIKE %s", searchModel.SearchGroup.SearchColumn, "'%"+r.helper.Util.TranslitToEng(searchModel.Query)+"%'")

	queryOrder := fmt.Sprintf("ORDER BY %s", searchModel.SearchGroup.LabelColumn)
	query := fmt.Sprintf("%s %s %s %s %s %s %s", querySelect, queryFrom, join, condition, conditionTranslitToRu, conditionTranslitToEng, queryOrder)

	rows, err := r.helper.DB.IDB(c).QueryContext(c, query)
	if err != nil {
		return err
	}

	err = r.helper.DB.DB.ScanRows(c, rows, &searchModel.SearchGroup.SearchElements)
	return err
}
