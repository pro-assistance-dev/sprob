package basehandler

import (
	"reflect"

	pluralize "github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
)

// EventKey — имя сущности для событий уведомлений (kebab-plural, как у роутов:
// «orders», «events»). Совпадает с ключом авто-CRUD (routing.getKey).
func EventKey[T Relationable]() string {
	key := reflect.TypeFor[T]().Name()
	return strcase.ToKebab(pluralize.NewClient().Plural(key))
}

// itemID — id сущности из поля `ID` (uuid.NullUUID с методом String). Пусто, если
// у модели нет id — тогда событие уйдёт без ItemID (не критично).
func itemID(item any) string {
	v := reflect.Indirect(reflect.ValueOf(item))
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return ""
	}
	f := v.FieldByName("ID")
	if !f.IsValid() {
		return ""
	}
	if s, ok := f.Interface().(interface{ String() string }); ok {
		return s.String()
	}
	return ""
}

// emitNotify публикует событие уведомления, если приёмник задан. payload — плоский
// снапшот сущности (для условий/шаблонов правил). Безопасен при nil-приёмнике.
func (h *Handler[T]) emitNotify(action, itemID string, payload map[string]any) {
	if h.helper == nil || h.helper.Notify == nil {
		return
	}
	h.helper.Notify.PublishNotify(EventKey[T](), action, itemID, "", payload)
}

// eventPayload — плоский снапшот сущности для события: id, имя, код/статус и т.п.
// Без рефлексии по всем полям — правила обычно фильтруют по name/status; при
// необходимости сущность может дать свой снапшот (отдельный интерфейс).
func eventPayload(item any) map[string]any {
	out := map[string]any{}
	v := reflect.Indirect(reflect.ValueOf(item))
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return out
	}
	if id := itemID(item); id != "" {
		out["id"] = id
	}
	for _, field := range []string{"Name", "Title", "Code", "Status"} {
		f := v.FieldByName(field)
		if f.IsValid() && f.Kind() == reflect.String {
			key := strcase.ToLowerCamel(field)
			out[key] = f.String()
		}
	}
	return out
}
