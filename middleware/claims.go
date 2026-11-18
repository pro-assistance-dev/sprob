package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pro-assistance-dev/sprob/helpers/token"
)

type (
	Claim  string
	Claims []Claim
)

const (
	ClaimUserID    Claim = "user_id"
	ClaimDomainIDS Claim = "domain_ids"
)

var ClaimsSlice = Claims{ClaimUserID, ClaimDomainIDS}

func (item Claim) Inject(r *http.Request, h *token.Token) error {
	d, err := h.ExtractTokenMetadata(r, item)
	if err != nil {
		return err
	}
	*r = *r.WithContext(context.WithValue(r.Context(), item.String(), d)) //nolint:all
	// ctx = context.WithValue(ctx, claim, d)
	return nil
}

func (items Claims) Inject(r *http.Request, h *token.Token) (err error) {
	for i := range items {
		err = items[i].Inject(r, h)
		if err != nil {
			break
		}
	}
	return err
}

func (item Claim) String() string {
	return string(item)
}

func (item Claim) Split() []string {
	return strings.Split(item.String(), ",")
}

func (item Claim) FromContext(ctx context.Context) string {
	return ctx.Value(item.String()).(string)
}

func (item Claim) FromContextSlice(ctx context.Context) []string {
	return strings.Split(item.FromContext(ctx), ",")
}

// InjectExternal — подложить claims БЕЗ JWT.
//
// ⚠️ ЗАЧЕМ. Домен (forms/survey/…) читает `user_id` и `domain_ids` из
// контекста запроса. `Inject` выше берёт их из JWT — и это единственный
// путь, который пакет давал. Проект с СВОЕЙ авторизацией (cookie-сессия,
// API-ключ, Basic) вынужден писать собственную мидлварь, повторяя ключи
// и типы пакета. Такую прослойку пришлось написать для «Средневековки»
// (`sred/sprobauth`) — и её же пришлось бы повторить в следующем проекте.
//
// Теперь: проект вызывает `InjectExternal` и передаёт значения сам.
//
// ⚠️ КЛЮЧИ — СТРОКИ ("user_id", "domain_ids"), а не типизированные: пакет
// читает их через `ctx.Value(claim.String())`. Своя типизированная обёртка
// здесь НЕ работает — спрячет значение от читателей.
//
// ⚠️ `domainID` — идентификатор тенанта. У проектов с одним тенантом это
// константа из env, но передаётся ОТ ПРОЕКТА, а не берётся из запроса:
// принимать его из query/header значило бы дать любому вошедшему доступ к
// данным любого домена.
//
// Пример:
//
//	api := r.Group("/api", myAuth.Require(), middleware.InjectExternal(userID, domainID))
func InjectExternal(userID, domainID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		//nolint:staticcheck // ключ задан пакетом, заменить на типизированный нельзя
		ctx = context.WithValue(ctx, ClaimUserID.String(), userID)
		//nolint:staticcheck // то же
		ctx = context.WithValue(ctx, ClaimDomainIDS.String(), domainID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
