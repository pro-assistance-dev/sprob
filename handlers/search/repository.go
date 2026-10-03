package search

import (
	"context"
	"fmt"

	"github.com/pro-assistance-dev/sprob/models"
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
	g := searchModel.SearchGroup

	// ⚠️ ЗНАЧЕНИЯ ПОИСКА — ПАРАМЕТРАМИ (`?`), не конкатенацией: `Query` приходит
	// от клиента, и склейка его в строку давала SQL-инъекцию (апостроф ломал
	// запрос). Колонки/таблица берутся из конфигурации группы поиска (не из
	// тела запроса) — оставляем как идентификаторы.
	searchExpr := fmt.Sprintf(
		"replace(regexp_replace(%s, '[^а-яА-Яa-zA-Z0-9. ]', '', 'g'), ' ', '')",
		g.SearchColumn,
	)

	query := fmt.Sprintf(
		"SELECT %s.%s as value, substring(%s for 40) as label FROM %s WHERE %s ILIKE ? OR %s ILIKE ? OR %s ILIKE ? ORDER BY %s",
		g.Table, g.ValueColumn, g.LabelColumn, g.Table, searchExpr, searchExpr, searchExpr, g.LabelColumn,
	)

	rows, err := r.helper.DB.IDB(c).QueryContext(
		c, query,
		"%"+searchModel.Query+"%",
		"%"+r.helper.Util.TranslitToRu(searchModel.Query)+"%",
		"%"+r.helper.Util.TranslitToEng(searchModel.Query)+"%",
	)
	if err != nil {
		return err
	}

	err = r.helper.DB.DB.ScanRows(c, rows, &searchModel.SearchGroup.SearchElements)
	return err
}
