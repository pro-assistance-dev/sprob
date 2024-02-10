package tree

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type TreeCreator struct {
	ID         *string
	treeModels TreeModels
}

func NewTreeCreator(c *gin.Context) (*TreeCreator, error) {
	models, err := createTreeModels(c)
	if err != nil {
		return nil, err
	}
	return &TreeCreator{treeModels: models}, err
}

func createTreeModels(c *gin.Context) (TreeModels, error) {
	treeModels := make(TreeModels, 0)
	if c.Query("treeModel") == "" {
		return nil, nil
	}
	for _, arg := range c.QueryArray("treeModel") {
		treeModel, err := parseJSONToTreeModel(arg)
		if err != nil {
			return nil, err
		}
		treeModels = append(treeModels, &treeModel)
	}

	return treeModels, nil
}

func (i *TreeCreator) CreateTree(query *bun.SelectQuery) {
	if len(i.treeModels) != 0 {
		for _, tree := range i.treeModels {
			fmt.Println(tree.getTableAndCols())
			query = query.ColumnExpr(tree.getTableAndCols())
		}
		return
	}
}

// CreateOrder method
func (items TreeModels) CreateTree(query *bun.SelectQuery) {
	if len(items) != 0 {
		for _, tree := range items {
			query = query.OrderExpr(tree.getTableAndCols())
		}
		return
	}
}
