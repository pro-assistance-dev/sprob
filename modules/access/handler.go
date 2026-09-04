package access

import (
	"net/http"
	"os"
	"strconv"

	"github.com/pro-assistance-dev/sprob/modules/access/models"

	"github.com/gin-gonic/gin"
	"github.com/pro-assistance-dev/sprob/helper"
)

// Handler - endpoint'ы этапов 2-3: матрица доступа и журнал корректуры.
type Handler struct {
	helper *helper.Helper
	matrix *MatrixStore
}

func NewHandler(h *helper.Helper, matrix *MatrixStore) *Handler {
	return &Handler{helper: h, matrix: matrix}
}

// MyAccessRow - строка эффективного доступа текущего пользователя.
type MyAccessRow struct {
	Entity string `json:"entity"`
	Field  string `json:"field"`
	Access string `json:"access"` // '' / R / W / M
	Owner  string `json:"owner,omitempty"`
}

// MyAccess возвращает эффективную матрицу доступа текущего пользователя
// (для ролевого UI: скрытие/блокировка полей и вкладок).
//
// @Summary Эффективная матрица доступа текущего пользователя
// @Description Роли и права (R/W/M) по полям всех сущностей — источник ролевого UI на клиенте.
// @Tags access
// @Produce json
// @Success 200 {object} map[string]interface{} "roles + items (MyAccessRow)"
// @Router /access-matrix/my [get]
func (h *Handler) MyAccess(c *gin.Context) {
	uc := RolesFromRequest(c.Request.Context(), h.helper, c.Request)

	// Матрица кэшируется с TTL 60s и подгружается лениво при первом обращении
	// (middleware enforcement). Этот endpoint — обычно ПЕРВЫЙ запрос клиента после
	// входа, поэтому принудительно загружаем кэш, иначе access будет пустым для
	// всех полей и ролевой UI на клиенте скроет/покажет не то.
	h.matrix.ensureLoaded(c.Request.Context())

	rows := make([]MyAccessRow, 0)
	for _, entity := range Entities() {
		for field, fi := range entity.Fields {
			access := ""
			if hasRole(uc.Roles, AdminRole) {
				access = AccessWrite
			} else {
				for _, role := range uc.Roles {
					a := h.matrix.accessFor(entity.Key, field, role)
					if a == AccessRead || a == AccessWrite || a == AccessMandatory {
						if access == "" || (a == AccessRead && access != "") {
							access = a
						}
						if a == AccessWrite || a == AccessMandatory {
							access = a
							break
						}
					}
				}
			}
			owner := FieldOwner(entity.Key, field)
			if fi.Owner != "" {
				owner = fi.Owner
			}
			rows = append(rows, MyAccessRow{
				Entity: entity.Key, Field: field, Access: access, Owner: owner,
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"roles": uc.Roles, "items": rows})
}

// AuditLog возвращает записи журнала корректуры.
// Фильтры: entity, objectId, operation, userName (ILIKE), code (ILIKE),
// dateFrom, dateTo; пагинация: offset, limit (макс. 500).
//
// @Summary Журнал корректуры: список записей
// @Description Кто/что/когда менял: фильтры (entity, objectId, operation, userName, code, даты) + пагинация.
// @Tags access
// @Produce json
// @Param entity query string false "сущность (напр. rooms)"
// @Param objectId query string false "id объекта"
// @Param operation query string false "create|update|delete"
// @Param userName query string false "ILIKE-фильтр"
// @Param code query string false "ILIKE-фильтр по инв. коду"
// @Param dateFrom query string false "с даты"
// @Param dateTo query string false "по дату"
// @Param offset query int false "сдвиг"
// @Param limit query int false "лимит (макс. 500)"
// @Success 200 {object} map[string]interface{} "items + count"
// @Router /audit-log-query [get]
func (h *Handler) AuditLog(c *gin.Context) {
	uc := RolesFromRequest(c.Request.Context(), h.helper, c.Request)
	// При включённом enforcement журнал доступен только пользователям с ролями FM.
	if m := h.matrix; m != nil && os.Getenv("ACCESS_ENFORCE") == "true" && len(uc.Roles) == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "нет ролей FM"})
		return
	}

	entity := c.Query("entity")
	objectID := c.Query("objectId")
	operation := c.Query("operation")
	userName := c.Query("userName")
	code := c.Query("code")
	limit, _ := parseIntDefault(c.Query("limit"), 100)
	if limit > 500 {
		limit = 500
	}
	offset, _ := parseIntDefault(c.Query("offset"), 0)

	q := h.helper.DB.IDB(c.Request.Context()).NewSelect().Model(&models.AuditLog{})
	if entity != "" {
		q = q.Where("entity = ?", entity)
	}
	if objectID != "" {
		q = q.Where("object_id = ?", objectID)
	}
	if operation != "" {
		q = q.Where("operation = ?", operation)
	}
	if userName != "" {
		q = q.Where("user_name ILIKE ?", "%"+userName+"%")
	}
	if code != "" {
		q = q.Where("code ILIKE ?", "%"+code+"%")
	}
	if df := c.Query("dateFrom"); df != "" {
		q = q.Where("created_at >= ?", df)
	}
	if dt := c.Query("dateTo"); dt != "" {
		q = q.Where("created_at < ?", dt)
	}
	q = q.Order("created_at DESC").Limit(limit).Offset(offset)

	var logs models.AuditLogs
	count, err := q.ScanAndCount(c.Request.Context(), &logs)
	if err != nil {
		h.helper.HTTP.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": logs, "count": count})
}

// RolesList возвращает роли пользователя (для отладки и UI).
//
// @Summary Роли текущего пользователя
// @Description userId/userName/roles — для отладки и ролевого UI.
// @Tags access
// @Produce json
// @Success 200 {object} map[string]interface{} "userId + userName + roles"
// @Router /roles/my [get]
func (h *Handler) RolesList(c *gin.Context) {
	uc := RolesFromRequest(c.Request.Context(), h.helper, c.Request)
	c.JSON(http.StatusOK, gin.H{"userId": uc.UserID, "userName": uc.UserName, "roles": uc.Roles})
}

func parseIntDefault(s string, def int) (int, bool) {
	if s == "" {
		return def, false
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def, false
	}
	return v, true
}
