package tree

import (
	"fmt"
	"log"
	"testing"

	"github.com/pro-assistance-dev/sprob/config"
	"github.com/pro-assistance-dev/sprob/helpers/db"
	"github.com/pro-assistance-dev/sprob/helpers/project"
	// "github.com/pro-assistance-dev/sprob/helpers/sql/tree/mocks"
)

func prepare() *db.DB {
	conf, err := config.LoadTestConfig()
	if err != nil {
		log.Fatal(err)
	}
	p := project.NewProject(&conf.Project)
	p.InitSchemas()

	db := db.NewDB(conf.DB)

	return db
}

var tree = TreeModel{Model: "form", Cols: []string{"id", "name"}}

func TestGetTableAndCols(t *testing.T) {
	db := prepare()
	t.Run("CreateTree", func(t *testing.T) {
		selectQuery := db.DB.NewSelect()
		// .Model(mocks.Form{})
		tree.CreateTree(selectQuery)
		fmt.Println(selectQuery)
	})
}
