// Package testkit — лёгкий тестовый харнесс для кода на sprob (Т7 TECH_DEBT).
//
// Позволяет писать тесты в sprob и в серверах-потребителях БЕЗ поднятого
// Postgres и БЕЗ глобала basehandler.Helper:
//
//	h := testkit.NewSQLiteHelper(t, (*models.News)(nil))
//	baseR.InitR[models.News](api, baseR.WithHelper(h), baseR.WithKey("news"))
//	... httptest-запросы ...
//
// Зависимости собраны аддитивно: sqlite in-memory (bun + modernc), HTTP-хелпер,
// тестовый JWT (helper.Token с фиксированным секретом — см. Token).
//
// Анализ: archive/analysis-sprob-di-2026-09-03.md, analysis-sprob-testability-2026-09-03.md.
package testkit

import (
	"context"
	dbsql "database/sql"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/pro-assistance-dev/sprob/config"
	dbhelper "github.com/pro-assistance-dev/sprob/helpers/db"
	httphelper "github.com/pro-assistance-dev/sprob/helpers/http"
	sqlhelper "github.com/pro-assistance-dev/sprob/helpers/sql"
	tokenhelper "github.com/pro-assistance-dev/sprob/helpers/token"
	"github.com/pro-assistance-dev/sprob/helper"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	_ "modernc.org/sqlite"
)

// TestSecret — фиксированный секрет тестового токена (валиден и для
// middleware.InjectClaims: helper.Token создаётся с этим же секретом).
const TestSecret = "test-secret"

// NewSQLiteHelper — sqlite in-memory + bun; для переданных моделей создаёт
// таблицы (bun.NewCreateTable). Возвращает helper.Helper с DB/SQL/HTTP/Token —
// пригоден для авто-CRUD (routing.WithHelper), репозиториев и handler-тестов.
// Каждый вызов получает отдельную in-memory БД (можно t.Parallel()).
func NewSQLiteHelper(t testing.TB, models ...any) *helper.Helper {
	t.Helper()
	sqldb, err := dbsql.Open("sqlite", fmt.Sprintf("file:testkit-%s?mode=memory&cache=shared", uuid.NewString()))
	if err != nil {
		t.Fatalf("testkit: open sqlite: %v", err)
	}
	sqldb.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqldb.Close() })

	bdb := bun.NewDB(sqldb, sqlitedialect.New())
	ctx := context.Background()
	for _, m := range models {
		if m == nil {
			continue
		}
		if _, err := bdb.NewCreateTable().Model(m).IfNotExists().Exec(ctx); err != nil {
			t.Fatalf("testkit: create table %T: %v", m, err)
		}
	}

	return &helper.Helper{
		DB:   &dbhelper.DB{DB: bdb},
		SQL:  sqlhelper.NewSQL(),
		HTTP: &httphelper.HTTP{}, // HandleError не требует конфига
		Token: tokenhelper.NewToken(config.Token{
			TokenSecret:        TestSecret,
			TokenAccessMinutes: 60,
			TokenRefreshHours:  24,
		}),
	}
}

// CreateTable — создать таблицу модели в уже созданном хелпере
// (модели, добавленные после NewSQLiteHelper).
func CreateTable(t testing.TB, h *helper.Helper, models ...any) {
	t.Helper()
	ctx := context.Background()
	for _, m := range models {
		if _, err := h.DB.IDB(ctx).NewCreateTable().Model(m).IfNotExists().Exec(ctx); err != nil {
			t.Fatalf("testkit: create table %T: %v", m, err)
		}
	}
}

// claims — JWTClaimsSetter для подписи токена с нужными claims.
type claims map[string]any

func (c claims) SetJWTClaimsMap(m map[string]any) {
	for k, v := range c {
		m[k] = v
	}
}

// Token — подписанный JWT для защищённых роутов (middleware.InjectClaims):
// возвращает access-token c user_id и domain_ids. Секрет — TestSecret,
// тот же, что у helper.Token из NewSQLiteHelper.
func Token(h *helper.Helper, userID string, domainIDs ...string) string {
	if h.Token == nil {
		h.Token = tokenhelper.NewToken(config.Token{
			TokenSecret:        TestSecret,
			TokenAccessMinutes: 60,
			TokenRefreshHours:  24,
		})
	}
	c := claims{"user_id": userID, "domain_ids": domainIDs}
	td, err := h.Token.CreateToken(c)
	if err != nil {
		panic(fmt.Sprintf("testkit: sign token: %v", err))
	}
	return td.AccessToken
}
