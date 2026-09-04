package access

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/pro-assistance-dev/sprob/modules/access/models"

	"github.com/pro-assistance-dev/sprob/helper"
)

// Access values.
const (
	AccessNone      = ""
	AccessRead      = "R"
	AccessWrite     = "W"
	AccessMandatory = "M"
)

// MatrixRow - строка матрицы: доступ роли к полю сущности.
type MatrixRow struct {
	Entity   string
	Field    string
	RoleCode string
	Access   string
	Owner    string
}

// MatrixStore - кэш матрицы доступа (загружается из access_matrix).
type MatrixStore struct {
	helper *helper.Helper
	mu     sync.RWMutex
	// entity -> field -> role -> access
	matrix map[string]map[string]map[string]string
	// entity -> field -> owner
	owners map[string]map[string]string
	loaded time.Time
	ttl    time.Duration
}

func NewMatrixStore(h *helper.Helper) *MatrixStore {
	return &MatrixStore{
		helper: h,
		matrix: map[string]map[string]map[string]string{},
		owners: map[string]map[string]string{},
		ttl:    60 * time.Second,
	}
}

// Load - полная перезагрузка матрицы из БД.
func (m *MatrixStore) Load(ctx context.Context) error {
	var rows []models.AccessMatrix
	err := m.helper.DB.IDB(ctx).NewSelect().Model(&rows).Scan(ctx)
	if err != nil {
		return err
	}

	matrix := map[string]map[string]map[string]string{}
	owners := map[string]map[string]string{}
	for _, r := range rows {
		if matrix[r.Entity] == nil {
			matrix[r.Entity] = map[string]map[string]string{}
		}
		if matrix[r.Entity][r.Field] == nil {
			matrix[r.Entity][r.Field] = map[string]string{}
		}
		matrix[r.Entity][r.Field][r.RoleCode] = r.Access
		if r.OwnerRole != "" {
			if owners[r.Entity] == nil {
				owners[r.Entity] = map[string]string{}
			}
			owners[r.Entity][r.Field] = r.OwnerRole
		}
	}

	m.mu.Lock()
	m.matrix = matrix
	m.owners = owners
	m.loaded = time.Now()
	m.mu.Unlock()
	return nil
}

func (m *MatrixStore) ensureLoaded(ctx context.Context) {
	m.mu.RLock()
	stale := m.loaded.IsZero() || time.Since(m.loaded) > m.ttl
	m.mu.RUnlock()
	if stale {
		_ = m.Load(ctx)
	}
}

// EnsureLoaded принудительно загружает матрицу, если кэш пуст или протух.
// Экспортирован для пакетов, которым нужна актуальная матрица (например, mytasks).
func (m *MatrixStore) EnsureLoaded(ctx context.Context) {
	m.ensureLoaded(ctx)
}

func (m *MatrixStore) accessFor(entity, field, role string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.matrix[entity] == nil || m.matrix[entity][field] == nil {
		return ""
	}
	return m.matrix[entity][field][role]
}

// AccessFor возвращает доступ роли к полю сущности (” / R / W / M).
// Экспортирован для пакетов, которым нужен разбор матрицы (например, mytasks).
func (m *MatrixStore) AccessFor(entity, field, role string) string {
	return m.accessFor(entity, field, role)
}

// Owner возвращает роль-владельца поля ("" - не задана).
func (m *MatrixStore) Owner(entity, field string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.owners[entity] == nil {
		return ""
	}
	return m.owners[entity][field]
}

func hasRole(roles []string, role string) bool {
	for _, r := range roles {
		if strings.EqualFold(r, role) {
			return true
		}
	}
	return false
}

// CanRead - есть ли у кого-то из ролей право R/W/M на поле.
func (m *MatrixStore) CanRead(ctx context.Context, entity, field string, roles []string) bool {
	m.ensureLoaded(ctx)
	if hasRole(roles, AdminRole) {
		return true
	}
	for _, role := range roles {
		a := m.accessFor(entity, field, role)
		if a == AccessRead || a == AccessWrite || a == AccessMandatory {
			return true
		}
	}
	// Поля, отсутствующие в матрице, читаются всеми (обратная совместимость).
	return m.accessFor(entity, field, "") == "" && !m.hasAnyRow(entity, field)
}

// CanWrite - есть ли право W/M на поле.
func (m *MatrixStore) CanWrite(ctx context.Context, entity, field string, roles []string) bool {
	m.ensureLoaded(ctx)
	if hasRole(roles, AdminRole) {
		return true
	}
	for _, role := range roles {
		a := m.accessFor(entity, field, role)
		if a == AccessWrite || a == AccessMandatory {
			return true
		}
	}
	return false
}

// CanWriteEntity - есть ли право W/M хотя бы на одно поле сущности (для ярусов 2/3).
func (m *MatrixStore) CanWriteEntity(ctx context.Context, entity string, roles []string) bool {
	m.ensureLoaded(ctx)
	if hasRole(roles, AdminRole) {
		return true
	}
	ent := GetEntity(entity)
	if ent == nil {
		return true
	}
	for field := range ent.Fields {
		if m.CanWrite(ctx, entity, field, roles) {
			return true
		}
	}
	return false
}

// CanDelete - право на удаление: W на сущности (или R00).
func (m *MatrixStore) CanDelete(ctx context.Context, entity string, roles []string) bool {
	if hasRole(roles, AdminRole) {
		return true
	}
	ent := GetEntity(entity)
	if ent == nil {
		return true
	}
	if ent.EntityLevel {
		return m.CanWriteEntity(ctx, entity, roles)
	}
	for field := range ent.Fields {
		if m.CanWrite(ctx, entity, field, roles) {
			return true
		}
	}
	return false
}

func (m *MatrixStore) hasAnyRow(entity, field string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.matrix[entity] != nil && m.matrix[entity][field] != nil
}

// FieldOwner - владелец поля из реестра (эталон) либо из матрицы.
func FieldOwner(entity, field string) string {
	ent := GetEntity(entity)
	if ent != nil {
		if fi, ok := ent.Fields[field]; ok {
			return fi.Owner
		}
	}
	return ""
}
