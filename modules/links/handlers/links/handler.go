package links

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pro-assistance-dev/sprob/helper"
	"github.com/pro-assistance-dev/sprob/modules/links/models"
	"github.com/pro-assistance-dev/sprob/modules/links/qr"
)

// Новая batch-модель для пакетного выпуска («до 6000 QR»): одним запросом —
// целый этаж/здание/отделение. Идемпотентно по (app, targetType, targetId).
type ResolverFunc func(ctx context.Context, link *models.Link) (string, error)

type Handler struct {
	helper   *helper.Helper
	svc      *Service
	resolver qr.Resolver
}

// Service — доменная логика ссылок (коды, домены, пакетный выпуск).
type Service struct {
	helper   *helper.Helper
	resolver qr.Resolver
}

func New(h *helper.Helper, opts ...Option) *Handler {
	o := options{}
	for _, opt := range opts {
		opt(&o)
	}
	svc := &Service{helper: h, resolver: o.resolver}
	return &Handler{helper: h, svc: svc, resolver: o.resolver}
}

type options struct {
	resolver qr.Resolver
}

type Option func(*options)

// WithResolver — как собирать цель ссылки (адаптер сервиса-потребителя).
func WithResolver(r qr.Resolver) Option {
	return func(o *options) { o.resolver = r }
}

// ResolveTarget — абсолютный URL цели или ошибка домена (ErrBadTarget и т.п.).
func (h *Handler) ResolveTarget(ctx context.Context, link *models.Link) (string, error) {
	return h.resolver.Resolve(ctx, link)
}

// Ensure — вернуть существующую ссылку цели или создать новую (идемпотентно).
func (h *Handler) Ensure(ctx context.Context, link *models.Link) (*models.Link, error) {
	db := h.helper.DB.IDB(ctx)
	existing := &models.Link{}
	err := db.NewSelect().Model(existing).
		Where("app = ?", link.App).
		Where("target_type = ?", link.TargetType).
		Where("target_id = ?", link.TargetID).
		Limit(1).
		Scan(ctx)
	if err == nil {
		return existing, nil
	}
	if err = h.Create(ctx, link); err != nil {
		return nil, err
	}
	return link, nil
}

// Create — вставка с генерацией короткого уникального кода (ретраи на коллизию).
//
// ⚠️ Прямой SQL, не модель: Bun кэширует схему структуры, и авто-подстановка
// значения для `link_code` молча давала ПУСТОЙ код (ссылка //r/ и duplicate key
// на второй ссылке). Column/Value/ExcludeColumn/Safe это не лечат. Код — это
// идентификатор наклейки, здесь ручной SQL оправдан.
func (h *Handler) Create(ctx context.Context, link *models.Link) error {
	if link.RedirectKind == 0 {
		link.RedirectKind = 302
	}
	if !link.Enabled {
		// JSON без поля даёт false, а это значило бы «наклейка не работает»
		// сразу после выпуска: новая ссылка всегда включена.
		link.Enabled = true
	}
	return h.InsertGenerated(ctx, link)
}

// InsertGenerated — вставка с кодом, сгенерированным на уровне SQL.
//
// ⚠️ Единая точка вставки ссылки: и `Create`, и авто-CRUD (`POST /link-registry`)
// идут сюда. baseR зовёт `bun.NewInsert().Model(link)`, а Model-вставка
// подставляет ПУСТОЙ `link_code` (схема кэширует колонку) — ссылка выходила
// с кодом `""` (печаталась наклейка `//r/`). Теперь: код генерится в цикле до
// успешной вставки, а у модели поле `Code` помечено `scanonly` (не пишется).
func (s *Service) InsertGenerated(ctx context.Context, link *models.Link) error {
	h := &Handler{helper: s.helper, svc: s, resolver: s.resolver}
	return h.InsertGenerated(ctx, link)
}

// InsertGenerated — см. Service.InsertGenerated.
func (h *Handler) InsertGenerated(ctx context.Context, link *models.Link) error {
	db := h.helper.DB.IDB(ctx)
	for attempt := 0; attempt < 8; attempt++ {
		code, err := qr.NewCode()
		if err != nil {
			return err
		}
		link.Code = code
		row := db.QueryRowContext(ctx, insertLinkSQL,
			code, link.Title, link.App, link.TargetType, link.TargetID, link.ExternalURL, link.Path,
			link.Enabled, link.RedirectKind, link.ExpiresAt, link.Note)
		err = row.Scan(&link.ID, &link.Hits, &link.CreatedAt, &link.UpdatedAt)
		if err == nil {
			return nil
		}
		if !isUniqueViolation(err) {
			return err
		}
	}
	return ErrCodeCollision
}

// CreateEntry — POST /api/link-registry: создать ссылку с коротким кодом.
//
// Отличается от baseR-хендлера только вставкой: `InsertGenerated` (а не
// `NewInsert().Model`), иначе `link_code` остаётся пустым.
func (h *Handler) CreateEntry(c *gin.Context) {
	link := &models.Link{}
	if _, err := h.helper.HTTP.GetForm(c, link); err != nil {
		h.helper.HTTP.HandleError(c, err)
		return
	}
	if err := h.Create(c.Request.Context(), link); err != nil {
		h.helper.HTTP.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, link)
}

// DefaultDomain — активный домен оберток.
func (h *Handler) DefaultDomain(ctx context.Context) (*models.LinkDomain, error) {
	domain := &models.LinkDomain{}
	err := h.helper.DB.IDB(ctx).NewSelect().Model(domain).
		Where("is_default").
		Limit(1).
		Scan(ctx)
	return domain, err
}

// ShortURL — обёртка /r/<code> на активном домене.
func (h *Handler) ShortURL(ctx context.Context, code string) (string, error) {
	domain, err := h.DefaultDomain(ctx)
	if err != nil {
		return "", models.ErrNoDomain
	}
	return domain.Scheme + "://" + domain.Host + "/r/" + code, nil
}

// PurgeHits — ретеншен журнала переходов (LINK_HIT_RETENTION_DAYS, дефолт 180).
func (h *Handler) PurgeHits(ctx context.Context, days int) (int64, error) {
	if days <= 0 {
		days = 180
	}
	res, err := h.helper.DB.IDB(ctx).NewDelete().Model((*models.LinkHit)(nil)).
		Where("created_at < now() - (? || ' days')::interval", days).
		Exec(ctx)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

// now — единственная точка «текущего времени» для проверки срока (тесты).
func now() time.Time {
	return time.Now()
}

// FindByCode — ссылка по короткому коду (для публичного /r/:code).
func (h *Handler) FindByCode(ctx context.Context, code string) (*models.Link, error) {
	link := &models.Link{}
	err := h.helper.DB.IDB(ctx).NewSelect().Model(link).Where("link_code = ?", code).Scan(ctx)
	if err != nil {
		return nil, models.ErrNotFound
	}
	return link, nil
}

// Hit — журнал перехода (вызывается асинхронно из redirect-обработчика).
func (h *Handler) Hit(ctx context.Context, link *models.Link, ip, ua, referer string) error {
	db := h.helper.DB.IDB(ctx)
	_, err := db.NewInsert().Model(&models.LinkHit{
		LinkID: link.ID, Code: link.Code, IP: ip, UserAgent: ua, Referer: referer,
	}).Exec(ctx)
	if err != nil {
		return err
	}
	_, err = db.NewUpdate().Model((*models.Link)(nil)).
		Set("hits = hits + 1").
		Where("id = ?", link.ID).
		Exec(ctx)
	return err
}

// CheckState — можно ли вести по ссылке (включена, не истекла).
func CheckState(link *models.Link, now time.Time) error {
	if !link.Enabled {
		return models.ErrDisabled
	}
	if link.ExpiresAt != nil && now.After(*link.ExpiresAt) {
		return models.ErrExpired
	}
	return nil
}

// ── служебное ───────────────────────────────────────────────────────────────

var ErrCodeCollision = errCodeCollision{}

type errCodeCollision struct{}

func (errCodeCollision) Error() string { return "link code collision" }

// isUniqueViolation — pg-ошибка unique (код 23505) без жёсткой зависимости от драйвера.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	type sqlStater interface{ SQLState() string }
	if st, ok := err.(sqlStater); ok {
		return st.SQLState() == "23505"
	}
	return false
}

// UUIDString — id ссылки строкой (для экспорта/логов).
func UUIDString(id uuid.NullUUID) string {
	if !id.Valid {
		return ""
	}
	return id.UUID.String()
}
