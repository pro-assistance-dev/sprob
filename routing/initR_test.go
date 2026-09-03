package routing

import (
	"context"
	dbsql "database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pro-assistance-dev/sprob/handlers/basehandler"
	dbhelper "github.com/pro-assistance-dev/sprob/helpers/db"
	httphelper "github.com/pro-assistance-dev/sprob/helpers/http"
	sqlhelper "github.com/pro-assistance-dev/sprob/helpers/sql"
	"github.com/pro-assistance-dev/sprob/helper"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	_ "modernc.org/sqlite"
)

// Т7.1: авто-CRUD роуты монтируются через WithHelper(h) — БЕЗ глобала
// basehandler.Helper/SetHelper: полный цикл «gin + sqlite in-memory + httptest».
type routeItem struct {
	bun.BaseModel `bun:"route_test_items,alias:route_test_items"`
	ID            string `bun:"id,pk"`
	Name          string `bun:"name"`
}

func (routeItem) Relation(q *bun.SelectQuery) *bun.SelectQuery { return q }

func newRouteTestHelper(t *testing.T) *helper.Helper {
	t.Helper()
	sqldb, err := dbsql.Open("sqlite", "file:routing-test?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqldb.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqldb.Close() })

	bdb := bun.NewDB(sqldb, sqlitedialect.New())
	if _, err := bdb.NewCreateTable().Model((*routeItem)(nil)).Exec(context.Background()); err != nil {
		t.Fatalf("create table: %v", err)
	}
	return &helper.Helper{
		DB:   &dbhelper.DB{DB: bdb},
		SQL:  sqlhelper.NewSQL(),
		HTTP: &httphelper.HTTP{}, // HandleError не требует конфига
	}
}

func TestInitR_WithHelper_NoGlobal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newRouteTestHelper(t)
	if basehandler.Helper != nil {
		t.Fatal("тест предполагает, что глобал basehandler.Helper не выставлен")
	}

	router := gin.New()
	api := router.Group("/api")
	InitR[routeItem](api, WithKey("route-test-items"), WithHelper(h))

	// 1. Пустой список
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/route-test-items", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET список: code = %d, body = %s", w.Code, w.Body.String())
	}
	var first struct {
		Items []routeItem `json:"items"`
		Count int         `json:"count"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &first); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if first.Count != 0 || len(first.Items) != 0 {
		t.Fatalf("пустая таблица: count = %d, items = %d", first.Count, len(first.Items))
	}

	// 2. Вставка через DB и повторный GET — данные видны через роут
	ctx := context.Background()
	if _, err := h.DB.IDB(ctx).NewInsert().Model(&routeItem{ID: "11111111-1111-1111-1111-111111111111", Name: "Первый"}).Exec(ctx); err != nil {
		t.Fatalf("insert: %v", err)
	}
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/route-test-items", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET список (после insert): code = %d", w.Code)
	}
	var second struct {
		Items []routeItem `json:"items"`
		Count int         `json:"count"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &second); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if second.Count != 1 || len(second.Items) != 1 || second.Items[0].Name != "Первый" {
		t.Fatalf("после insert: count = %d, items = %+v", second.Count, second.Items)
	}

	// 3. Options (/options/:label/:value)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/route-test-items/options/name/id", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET options: code = %d, body = %s", w.Code, w.Body.String())
	}
	var opts []struct {
		Label string `json:"label"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &opts); err != nil {
		t.Fatalf("decode options: %v", err)
	}
	if len(opts) != 1 || opts[0].Label != "Первый" || opts[0].Value != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("options = %+v", opts)
	}

	// 4. Get не-UUID → 404 (валидация параметра)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/route-test-items/not-a-uuid", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("GET /not-a-uuid: code = %d, want 404", w.Code)
	}
}
