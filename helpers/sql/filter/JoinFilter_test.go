package filter

import (
	"strings"
	"testing"

	"github.com/pro-assistance-dev/sprob/helpers/project"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

// Join-фильтр («Отдел» у работников): `worker` ← has-many → `workerPosition`.
//
// ⚠️ ЗАЧЕМ (жалоба владельца 08.11: «выбрал отдел — пустой список; другие
// фильтры тоже не работают»). Ветка собиралась, но НЕ применялась:
// `constructJoinV3` ждал колонку в `JoinIn` (её наполнял `mergeJoins`), а
// `mergeJoins` склеивал ВТОРОЙ фильтр по той же таблице в первый через общий
// `JoinIn` — из-за этого условие по значению (`Col` = `departmentId`) в SQL
// вообще не попадало, и список отдавался полный.

func joinSchemas() {
	worker := &project.Schema{
		NameTable:  "workers",
		NamePascal: "Worker",
		NameCamel:  "worker",
		FieldsMap: map[string]*project.SchemaField{
			"fullName": {NamePascal: "FullName", NameCamel: "fullName", NameCol: "full_name"},
		},
	}
	workerPosition := &project.Schema{
		NameTable:  "worker_positions",
		NamePascal: "WorkerPosition",
		NameCamel:  "workerPosition",
		FieldsMap: map[string]*project.SchemaField{
			"workerId":     {NamePascal: "WorkerID", NameCamel: "workerId", NameCol: "worker_id"},
			"departmentId": {NamePascal: "DepartmentID", NameCamel: "departmentId", NameCol: "department_id"},
			"positionId":   {NamePascal: "PositionID", NameCamel: "positionId", NameCol: "position_id"},
		},
	}
	project.SchemasLib = project.SchemasMap{"worker": worker, "workerPosition": workerPosition}
}

func renderJoin(t *testing.T, items FilterModels) string {
	t.Helper()
	joinSchemas()
	db := bun.NewDB(nil, pgdialect.New())
	sel := db.NewSelect()
	items.CreateFilter(sel)
	return sel.String()
}

func joinFilter(col string, set ...string) *FilterModel {
	return &FilterModel{
		Model:          "worker",
		Col:            col,
		Type:           JoinType,
		Operator:       In,
		Set:            set,
		JoinTableModel: "workerPosition",
	}
}

func TestJoinFilter_ValueInSQL(t *testing.T) {
	query := renderJoin(t, FilterModels{joinFilter("departmentId", "dep-1")})

	if !strings.Contains(query, `JOIN worker_positions`) {
		t.Fatalf("нет JOIN по join-модели: %s", query)
	}
	if !strings.Contains(query, `workers.id = worker_positions.worker_id`) {
		t.Fatalf("join-условие не по схеме (ожидали workers.id = worker_positions.worker_id): %s", query)
	}
	if !strings.Contains(query, `worker_positions.department_id`) || !strings.Contains(query, `'dep-1'`) {
		t.Fatalf("условие по значению не попало в SQL (фильтр «молчит»): %s", query)
	}
}

// Два join-фильтра по РАЗНЫМ колонкам одной таблицы: оба условия обязаны
// остаться, а `JOIN` — только один (иначе Postgres: «more than once»).
func TestJoinFilter_TwoColumnsOneJoin(t *testing.T) {
	query := renderJoin(t, FilterModels{
		joinFilter("departmentId", "dep-1"),
		joinFilter("positionId", "pos-1"),
	})

	if got := strings.Count(query, `JOIN worker_positions`); got != 1 {
		t.Fatalf("ожидали ровно один JOIN, получили %d: %s", got, query)
	}
	for _, want := range []string{`worker_positions.department_id`, `worker_positions.position_id`, `'dep-1'`, `'pos-1'`} {
		if !strings.Contains(query, want) {
			t.Fatalf("потеряно условие %s: %s", want, query)
		}
	}
}

func TestJoinFilter_BelongsToFKOnModel(t *testing.T) {
	worker := &project.Schema{
		NameTable:  "workers",
		NamePascal: "Worker",
		NameCamel:  "worker",
		FieldsMap: map[string]*project.SchemaField{
			"departmentId": {NamePascal: "DepartmentID", NameCamel: "departmentId", NameCol: "department_id"},
		},
	}
	department := &project.Schema{
		NameTable:  "departments",
		NamePascal: "Department",
		NameCamel:  "department",
		FieldsMap: map[string]*project.SchemaField{
			"name": {NamePascal: "Name", NameCamel: "name", NameCol: "name"},
		},
	}
	project.SchemasLib = project.SchemasMap{"worker": worker, "department": department}

	db := bun.NewDB(nil, pgdialect.New())
	sel := db.NewSelect()
	FilterModels{{
		Model: "worker", Col: "name", Type: JoinType, Operator: Eq, Value1: "Администрация",
		JoinTableModel: "department",
	}}.CreateFilter(sel)

	query := sel.String()
	if !strings.Contains(query, `workers.department_id = departments.id`) {
		t.Fatalf("belongs-to join должен идти по FK самой модели: %s", query)
	}
	if !strings.Contains(query, `departments.name IN ('')`) && !strings.Contains(query, `departments.name = 'Администрация'`) {
		t.Fatalf("условие по значению не попало в SQL: %s", query)
	}
}
