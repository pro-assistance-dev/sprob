package basehandler

import (
	"context"
	dbsql "database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	dbhelper "github.com/pro-assistance-dev/sprob/helpers/db"
	sqlhelper "github.com/pro-assistance-dev/sprob/helpers/sql"
	"github.com/pro-assistance-dev/sprob/helper"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	_ "modernc.org/sqlite"
)

// testItem — минимальная модель для CRUD-теста baseR (sqlite in-memory,
// без Postgres и без глобального Helper).
type testItem struct {
	bun.BaseModel `bun:"test_items,alias:test_items"`
	ID            string `bun:"id,pk"`
	Name          string `bun:"name"`
}

func (item testItem) Relation(q *bun.SelectQuery) *bun.SelectQuery { return q }

func newTestHelper(t *testing.T) *helper.Helper {
	t.Helper()
	sqldb, err := dbsql.Open("sqlite", "file:basehandler-test?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqldb.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqldb.Close() })

	bdb := bun.NewDB(sqldb, sqlitedialect.New())
	if _, err := bdb.NewCreateTable().Model((*testItem)(nil)).Exec(context.Background()); err != nil {
		t.Fatalf("create table: %v", err)
	}

	return &helper.Helper{DB: &dbhelper.DB{DB: bdb}, SQL: sqlhelper.NewSQL()}
}

// TestRepository_CRUD — полный цикл авто-CRUD baseR: create → get → options →
// getAll → update → delete (Т4.2 TECH_DEBT).
func TestRepository_CRUD(t *testing.T) {
	h := newTestHelper(t)
	repo := Repository[testItem]{helper: h}
	repo.relation = testItem{}.Relation // в проде задаётся InitH[T]()
	ctx := context.Background()

	// Create
	item := &testItem{ID: "item-1", Name: "Первый"}
	if err := repo.Create(ctx, item); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Get
	got, err := repo.Get(ctx, "item-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name != "Первый" {
		t.Errorf("Get().Name = %q, want %q", got.Name, "Первый")
	}

	// Options (label/value колонки)
	opts, err := repo.Options(ctx, "name", "id")
	if err != nil {
		t.Fatalf("Options: %v", err)
	}
	if len(opts) != 1 || opts[0].Label != "Первый" || opts[0].Value != "item-1" {
		t.Errorf("Options() = %+v, want [{Первый item-1}]", opts)
	}

	// GetAll (без FTSP в контексте — HandleQuery имеет nil-guard)
	all, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if all.Count != 1 || len(all.Items) != 1 {
		t.Errorf("GetAll() count=%d items=%d, want 1/1", all.Count, len(all.Items))
	}

	// Update
	item.Name = "Обновлённый"
	if err := repo.Update(ctx, item); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, err = repo.Get(ctx, "item-1")
	if err != nil {
		t.Fatalf("Get after Update: %v", err)
	}
	if got.Name != "Обновлённый" {
		t.Errorf("Get() after Update = %q, want %q", got.Name, "Обновлённый")
	}

	// Delete
	if err := repo.Delete(ctx, "item-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.Get(ctx, "item-1"); !errors.Is(err, dbsql.ErrNoRows) {
		t.Errorf("Get after Delete: err = %v, want sql.ErrNoRows", err)
	}
}

// TestRepository_GetUnknownID — Get несуществующей записи возвращает ошибку.
func TestRepository_GetUnknownID(t *testing.T) {
	h := newTestHelper(t)
	repo := Repository[testItem]{helper: h}
	repo.relation = testItem{}.Relation
	ctx := context.Background()

	if _, err := repo.Get(ctx, "no-such-id"); !errors.Is(err, dbsql.ErrNoRows) {
		t.Errorf("Get(unknown) err = %v, want sql.ErrNoRows", err)
	}
}

// TestHandler_Get_InvalidUUID — не-UUID в :id → 404 (не 500 от Postgres).
// v1.0.251: фикс 500 /api/events/{slug} в pros.
func TestHandler_Get_InvalidUUID(t *testing.T) {
	h := newTestHelper(t)
	router := gin.New()
	handler := Handler[testItem]{helper: h, S: InitS(Repository[testItem]{helper: h})}
	router.GET("/test-items/:id", handler.Get)

	req := httptest.NewRequest("GET", "/test-items/not-a-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("GET /test-items/not-a-uuid = %d, want 404", w.Code)
	}
}

// TestHandler_Delete_InvalidUUID — не-UUID в :id → 404.
func TestHandler_Delete_InvalidUUID(t *testing.T) {
	h := newTestHelper(t)
	router := gin.New()
	handler := Handler[testItem]{helper: h, S: InitS(Repository[testItem]{helper: h})}
	router.DELETE("/test-items/:id", handler.Delete)

	req := httptest.NewRequest("DELETE", "/test-items/not-a-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("DELETE /test-items/not-a-uuid = %d, want 404", w.Code)
	}
}

// ─── К2 (Т8): серверный поиск опций ──────────────────────────────────────

// searchableItem — модель с БЕЛЫМ СПИСКОМ полей поиска.
type searchableItem struct {
	bun.BaseModel `bun:"searchable_items,alias:searchable_items"`
	ID            string `bun:"id,pk"`
	Name          string `bun:"name"`
	FullName      string `bun:"full_name"`
}

func (item searchableItem) Relation(q *bun.SelectQuery) *bun.SelectQuery { return q }

// SearchColumns — только эти поля участвуют в поиске (белый список).
func (item searchableItem) SearchColumns() []string { return []string{"full_name", "name"} }

func newSearchableHelper(t *testing.T) *helper.Helper {
	t.Helper()
	sqldb, err := dbsql.Open("sqlite", "file:search-test?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqldb.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqldb.Close() })

	bdb := bun.NewDB(sqldb, sqlitedialect.New())
	if _, err := bdb.NewCreateTable().Model((*searchableItem)(nil)).Exec(context.Background()); err != nil {
		t.Fatalf("create table: %v", err)
	}
	return &helper.Helper{DB: &dbhelper.DB{DB: bdb}, SQL: sqlhelper.NewSQL()}
}

func seedSearchable(t *testing.T, repo Repository[searchableItem]) {
	t.Helper()
	ctx := context.Background()
	for _, it := range [][3]string{
		{"1", "Иванов Иван", "Сестра"},
		{"2", "Петров Пётр", "Врач"},
		{"3", "Сидоров Сидор", "Сестра"},
	} {
		item := &searchableItem{ID: it[0], FullName: it[1], Name: it[2]}
		if err := repo.Create(ctx, item); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
}

// TestRepository_OptionsSearch_ByWhitelist — поиск идёт по объявленным полям.
func TestRepository_OptionsSearch_ByWhitelist(t *testing.T) {
	h := newSearchableHelper(t)
	repo := Repository[searchableItem]{helper: h}
	repo.relation = searchableItem{}.Relation
	seedSearchable(t, repo)

	ctx := context.Background()

	// По подписи (full_name)
	got, err := repo.OptionsSearch(ctx, "full_name", "id", "Иван", 10)
	if err != nil {
		t.Fatalf("OptionsSearch: %v", err)
	}
	if len(got) != 1 || got[0].Value != "1" {
		t.Errorf("поиск 'Иван' = %+v, want [{1 Иванов Иван}]", got)
	}

	// По второму полю белого списка (name)
	got, err = repo.OptionsSearch(ctx, "full_name", "id", "Сестра", 10)
	if err != nil {
		t.Fatalf("OptionsSearch(2): %v", err)
	}
	if len(got) != 2 {
		t.Errorf("поиск 'Сестра' = %d шт, want 2", len(got))
	}

	// Пустой query — полный справочник (поведение как до К2)
	got, err = repo.OptionsSearch(ctx, "full_name", "id", "", 10)
	if err != nil {
		t.Fatalf("OptionsSearch(empty): %v", err)
	}
	if len(got) != 3 {
		t.Errorf("пустой query = %d шт, want 3", len(got))
	}
}

// TestRepository_OptionsSearch_Limit — лимит отсекает хвост и ограничен сверху.
func TestRepository_OptionsSearch_Limit(t *testing.T) {
	h := newSearchableHelper(t)
	repo := Repository[searchableItem]{helper: h}
	repo.relation = searchableItem{}.Relation
	seedSearchable(t, repo)
	ctx := context.Background()

	got, err := repo.OptionsSearch(ctx, "full_name", "id", "", 2)
	if err != nil {
		t.Fatalf("OptionsSearch: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("limit=2 → %d шт, want 2", len(got))
	}

	// Запрошенный сверх потолка лимит не ломает запрос (клампится).
	if _, err := repo.OptionsSearch(ctx, "full_name", "id", "", 100000); err != nil {
		t.Errorf("limit сверх потолка: %v", err)
	}
}

// TestRepository_OptionsSearch_NotSearchable — модель без SearchColumns → ошибка.
func TestRepository_OptionsSearch_NotSearchable(t *testing.T) {
	h := newTestHelper(t) // testItem без SearchColumns
	repo := Repository[testItem]{helper: h}
	repo.relation = testItem{}.Relation

	_, err := repo.OptionsSearch(context.Background(), "name", "id", "что-то", 10)
	if !errors.Is(err, ErrNotSearchable) {
		t.Errorf("err = %v, want ErrNotSearchable", err)
	}
}

// TestHandler_Options_NotSearchable400 — клиенту говорим 400, а не 500.
func TestHandler_Options_NotSearchable400(t *testing.T) {
	h := newTestHelper(t)
	router := gin.New()
	base := Repository[testItem]{helper: h}
	base.relation = testItem{}.Relation
	handler := Handler[testItem]{helper: h, S: InitS(base)}
	router.GET("/test-items/options/:label/:value", handler.Options)

	req := httptest.NewRequest("GET", "/test-items/options/name/id?query=abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("options?query= на не-searchable = %d, want 400", w.Code)
	}
}
