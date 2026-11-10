package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
)

// snapshotTables — таблица для снимка «до» по имени сущности ручки.
//
// ⚠️ Снимок берём из БАЗЫ, а не из тела запроса: тело содержит только то, что
// клиент отправлял (часто неполный объект), а откатывать нужно к реальному
// состоянию строки. Сущности, которых нет в таблице, просто не получат «было» —
// дельта по ним не строится, но журнал продолжает работать.
//
// ⚠️ Карта ОБЩАЯ для чтения/дельты; откат использует свою (`Options.RevertTables`),
// потому что откатывать можно не всё, что можно прочитать.
var snapshotTables = map[string]string{
	"workers":                "workers",
	"worker-positions":       "worker_positions",
	"worker-documents":       "worker_documents",
	"worker-document-fields": "worker_document_fields",
	"departments":            "departments",
	"positions":              "positions",
	"document-types":         "document_types",
	"department-workers":     "worker_positions",
	"educations":             "educations",
	"experiences":            "experiences",
	"regalias":               "regalias",
	"phones":                 "phones",
	"emails":                 "emails",
	"addresses":              "addresses",
	"users":                  "users",
	"humans":                 "humans",
	"contact-items":          "contact_items",
	"bonuses":                "bonuses",
	"medical-certificates":   "worker_documents",
	// map / portal (кладутся, чтобы общий журнал работал без правок проекта)
	"rooms":        "rooms",
	"buildings":    "buildings",
	"floors":       "floors",
	"polygons":     "polygons",
	"doctors":      "doctors",
	"news":         "news",
	"publications": "publications",
	"pages":        "pages",
}

// excludedColumns — служебные колонки, которые не попадают в снимок.
//
// ⚠️ Если их включить, откат переписывал бы историю (`updated_at` вернулся бы
// к прошлому значению и соврал о времени правки).
var excludedColumns = map[string]bool{
	"created_at": true, "updated_at": true, "created_by": true, "updated_by": true,
}

// objectIDFrom — id объекта: сначала из пути (`/:id`), потом из тела.
func objectIDFrom(c interface {
	Param(string) string
}, payload map[string]interface{},
) string {
	if s := c.Param("id"); s != "" {
		return s
	}
	if s, ok := payload["id"].(string); ok {
		return s
	}
	return ""
}

// snapshot читает строку «до» изменения.
//
// ⚠️ Ошибки не валят запрос: журнал — вспомогательная система. Нет таблицы/
// колонки/строки — просто не будет «было» (и откат по такой записи недоступен).
func (a *Audit) snapshot(ctx context.Context, entity, objectID string) map[string]interface{} {
	if objectID == "" {
		return nil
	}
	table, ok := snapshotTables[entity]
	if !ok {
		return nil
	}
	if _, err := uuid.Parse(objectID); err != nil {
		return nil
	}

	cols, err := a.scalarColumns(ctx, table)
	if err != nil || len(cols) == 0 {
		return nil
	}

	query := fmt.Sprintf(`SELECT jsonb_build_object(%s)::text AS snap FROM %q WHERE id = ?`,
		strings.Join(pairsForJSON(cols), ", "), table)
	// ⚠️ Сканируем в `string`, а не в `map`: bun превращал результат
	// `jsonb_build_object` в base64-строку, и снимок уезжал в журнал мусором
	// (08.11). Явный `::text` + собственный разбор JSON даёт ровно объект.
	snap := struct {
		Snap string `bun:"snap"`
	}{}
	if err := a.helper.DB.IDB(ctx).NewRaw(query, objectID).Scan(ctx, &snap); err != nil {
		log.Printf("[audit] snapshot %s/%s: %v", entity, objectID, err)
		return nil
	}
	if snap.Snap == "" {
		return nil
	}
	row := map[string]interface{}{}
	if err := json.Unmarshal([]byte(snap.Snap), &row); err != nil {
		log.Printf("[audit] snapshot decode %s/%s: %v", entity, objectID, err)
		return nil
	}
	return row
}

// scalarColumns — список колонок таблицы без служебных.
func (a *Audit) scalarColumns(ctx context.Context, table string) ([]string, error) {
	rows := make([]struct {
		ColumnName string `bun:"column_name"`
	}, 0)
	if err := a.helper.DB.IDB(ctx).NewRaw(`
		SELECT column_name FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = ?
		ORDER BY ordinal_position`, table).Scan(ctx, &rows); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		if !excludedColumns[r.ColumnName] {
			out = append(out, r.ColumnName)
		}
	}
	return out, nil
}

// pairsForJSON — `'col', "col"` для jsonb_build_object.
func pairsForJSON(cols []string) []string {
	out := make([]string, 0, len(cols))
	for _, c := range cols {
		out = append(out, fmt.Sprintf("'%s', %q", c, c))
	}
	return out
}

// encodeSnapshot — снимок в строку журнала (обрезаем, как и new_value).
func encodeSnapshot(snap map[string]interface{}) *string {
	if len(snap) == 0 {
		return nil
	}
	b, err := json.Marshal(snap)
	if err != nil {
		return nil
	}
	s := string(b)
	if len(s) > 4000 {
		s = s[:4000] + "..."
	}
	return &s
}
