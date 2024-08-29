package tree

import (
	"github.com/uptrace/bun"
)

// type Tree struct { //nolint:golint
// 	ID        *string
// 	TreeModel TreeModel
// }

// func NewTree(c *gin.Context) (*Tree, error) {
// 	models, err := createTreeModels(c)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &Tree{TreeModel: models}, err
// }
//
// func createTreeModels(c *gin.Context) (TreeModel, error) {
//
// 	if c.Query("treeModel") == "" {
// 		return nil, nil
// 	}
// 	for _, arg := range c.QueryArray("treeModel") {
// 		treeModel, err := parseJSONToTreeModel(arg)
// 		if err != nil {
// 			return nil, err
// 		}
// 		TreeModel = append(TreeModel, &treeModel)
// 	}
//
// 	return TreeModel, nil
// }

func (i *TreeModel) CreateTree(query *bun.SelectQuery) {
	expr := i.getTableAndCols()
	query.ColumnExpr(expr)
}

// CreateOrder method
// func (items Tree) CreateTree(query *bun.SelectQuery) {
// 	if len(items) != 0 {
// 		for _, tree := range items {
// 			query = query.OrderExpr(tree.getTableAndCols())
// 		}
// 		return
// 	}
// }
