package testkit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pro-assistance-dev/sprob/middleware"
	"github.com/uptrace/bun"
)

type kitItem struct {
	bun.BaseModel `bun:"kit_test_items,alias:kit_test_items"`
	ID            string `bun:"id,pk"`
	Name          string `bun:"name"`
}

func (kitItem) Relation(q *bun.SelectQuery) *bun.SelectQuery { return q }

// NewSQLiteHelper создаёт таблицы моделей и отдаёт рабочий helper.
func TestNewSQLiteHelper_CreatesTablesAndQueries(t *testing.T) {
	h := NewSQLiteHelper(t, (*kitItem)(nil))

	ctx := context.Background()
	if _, err := h.DB.IDB(ctx).NewInsert().Model(&kitItem{ID: "11111111-1111-1111-1111-111111111111", Name: "один"}).Exec(ctx); err != nil {
		t.Fatalf("insert: %v", err)
	}
	var got kitItem
	if err := h.DB.IDB(ctx).NewSelect().Model(&got).Where("id = ?", "11111111-1111-1111-1111-111111111111").Scan(ctx); err != nil {
		t.Fatalf("select: %v", err)
	}
	if got.Name != "один" {
		t.Errorf("Name = %q, want один", got.Name)
	}
}

// Token подписывается тем же секретом, что helper.Token — middleware
// InjectClaims должен его принять (проверяем подпись через ExtractTokenMetadata).
func TestToken_SignedWithHelperSecret(t *testing.T) {
	h := NewSQLiteHelper(t)
	tok := Token(h, "user-1", "domain-1")
	if tok == "" {
		t.Fatal("Token() вернул пустую строку")
	}

	// Валидация тем же ключом (как делает middleware.Claims.Inject).
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("token", tok)
	userID, err := h.Token.ExtractTokenMetadata(req, middleware.ClaimUserID)
	if err != nil {
		t.Fatalf("ExtractTokenMetadata: %v", err)
	}
	if userID != "user-1" {
		t.Errorf("user_id = %q, want user-1", userID)
	}
}

// Два хелпера изолированы: данные одного не видны в другом.
func TestNewSQLiteHelper_IsolatedDBs(t *testing.T) {
	h1 := NewSQLiteHelper(t, (*kitItem)(nil))
	h2 := NewSQLiteHelper(t, (*kitItem)(nil))

	ctx := context.Background()
	if _, err := h1.DB.IDB(ctx).NewInsert().Model(&kitItem{ID: "11111111-1111-1111-1111-111111111111", Name: "x"}).Exec(ctx); err != nil {
		t.Fatalf("insert h1: %v", err)
	}
	count, err := h2.DB.IDB(ctx).NewSelect().Model((*kitItem)(nil)).Count(ctx)
	if err != nil {
		t.Fatalf("count h2: %v", err)
	}
	if count != 0 {
		t.Errorf("h2 видит данные h1: count = %d", count)
	}
	_ = gin.New()
}
