package tree

import (
	"fmt"
	"os"
	"testing"

	"github.com/pro-assistance-dev/sprob/config"
	"github.com/pro-assistance-dev/sprob/helpers/db"
	"github.com/pro-assistance-dev/sprob/helpers/project"
	// "github.com/pro-assistance-dev/sprob/helpers/sql/tree/mocks"
)

// Интеграционный smoke: нужны локальный test.env и Postgres.
// В CI (sprob/.github/workflows/ci.yml) без SPROB_TEST_DB — пропускается.
func prepare(t *testing.T) *db.DB {
	t.Helper()
	if os.Getenv("SPROB_TEST_DB") == "" {
		t.Skip("интеграционный тест: нужна БД (SPROB_TEST_DB=1 + локальный test.env)")
	}
	conf, err := config.LoadTestConfig()
	if err != nil {
		t.Fatal(err)
	}
	p := project.NewProject(&conf.Project)
	p.InitSchemas()

	db := db.NewDB(conf.DB)

	return db
}

var tree = TreeModel{Model: "form", Cols: []string{"id", "name"}}

func TestGetTableAndCols(t *testing.T) {
	db := prepare(t)
	t.Run("CreateTree", func(t *testing.T) {
		selectQuery := db.DB.NewSelect()
		// .Model(mocks.Form{})
		tree.CreateTree(selectQuery)
		fmt.Println(selectQuery)
	})
}
