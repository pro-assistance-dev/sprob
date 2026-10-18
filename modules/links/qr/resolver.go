package qr

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/pro-assistance-dev/sprob/modules/links/models"
)

// Resolver собирает ЦЕЛЬ ссылки: (app, targetType) → абсолютный URL.
// Домен и маска живут в адаптере сервиса-потребителя, а не в ссылке:
// сменился домен — меняется резолвер/домен-запись, наклейки остаются.
type Resolver interface {
	// Resolve возвращает абсолютный URL цели. Пустой URL + nil error = цель
	// неизвестна этому резолверу (пробуем следующий).
	Resolve(ctx context.Context, link *models.Link) (string, error)
}

// StaticResolver — простая карта правил (удобно для внешних ссылок, тестов и
// сервисов без собственных сущностей). Ключ — "app/targetType".
type StaticResolver struct {
	Rules map[string]string
}

func (r StaticResolver) Resolve(_ context.Context, link *models.Link) (string, error) {
	if link.TargetType == "external" || link.ExternalURL != "" {
		if link.ExternalURL == "" {
			return "", models.ErrBadTarget
		}
		return link.ExternalURL, nil
	}
	if link.Path != "" && strings.HasPrefix(link.Path, "http") {
		return link.Path, nil
	}
	pattern, ok := r.Rules[link.App+"/"+link.TargetType]
	if !ok {
		return "", nil
	}
	return expand(pattern, link), nil
}

// expand подставляет {id}/{path} в шаблон цели.
func expand(pattern string, link *models.Link) string {
	out := strings.ReplaceAll(pattern, "{id}", link.TargetID)
	if link.Path != "" {
		out = strings.ReplaceAll(out, "{path}", strings.TrimPrefix(link.Path, "/"))
	} else {
		out = strings.ReplaceAll(out, "{path}", "")
	}
	return out
}

// Chain — первый резолвер, давший непустой URL.
type Chain []Resolver

func (c Chain) Resolve(ctx context.Context, link *models.Link) (string, error) {
	var lastErr error
	for _, r := range c {
		url, err := r.Resolve(ctx, link)
		if err != nil {
			if errors.Is(err, models.ErrBadTarget) {
				lastErr = err
				continue
			}
			return "", err
		}
		if url != "" {
			return url, nil
		}
	}
	if lastErr != nil {
		return "", lastErr
	}
	return "", fmt.Errorf("%w: %s/%s", models.ErrBadTarget, link.App, link.TargetType)
}
