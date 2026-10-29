// Package access — движок матрицы доступа (RACI) и журналов для проектов больницы:
// матрица ролей (roles/access_matrix/user_roles), enforcement на бэкенде
// (middleware), журнал корректуры (audit_log), журнал входа-выхода (auth_log),
// верификация JWT через JWKS keycloak (jwks.go).
//
// Реестр сущностей (что именно охраняем: entity → поля → роль-владелец) —
// ПРОЕКТНЫЙ: проект регистрирует свой список через Register() до создания middleware
// (см. rdkb/map: registry данных сущностей FM-системы; у каждого проекта свой домен).
//
// Значения access в access_matrix: ” (нет), R (чтение), W (запись),
// M (запись + обязательно для роли). Роль AdminRole (по умолчанию R00_ADMIN,
// переопределяется ACCESS_ADMIN_ROLE) имеет полный доступ ко всем полям.
package access

// FieldInfo описывает поле сущности: JSON-имя, роль-владелец, обязательность для владельца (M).
type FieldInfo struct {
	JSON      string
	Owner     string
	Mandatory bool
}

// Entity описывает сущность, участвующую в матрице доступа и аудите.
type Entity struct {
	Key   string // URL-ключ (baseR): rooms / room-engineerings / spr-object-statuses / ...
	Table string // таблица БД
	// EntityLevel - проверка записи на уровне сущности (ярус 2/3 и прочие листы книги):
	// достаточно права W на сущность (один владелец), полевая проверка только для rooms и spr_*.
	EntityLevel bool
	Fields      map[string]FieldInfo // json-имя поля -> владелец/обязательность
	Relations   map[string]string    // json-имя связи -> entity key (для маскирования вложенных объектов)
	// M2M — связи many-to-many: json-имя поля -> join-таблица (room_hazard_factors,
	// room_sout_classes, …). Такие поля перечислены и в Fields (их надо проверять на
	// запись), и в Relations (маскировать как вложенный объект), но физически лежат
	// НЕ в строке сущности. Без этого списка middleware не может прочитать «как было»
	// и считает связь изменённой на каждом сохранении (баг 29.10: ложный 403
	// «нет права записи на поля: hazardFactors,soutClasses»).
	M2M map[string]M2MRelation
}

// M2MRelation описывает join-таблицу связи many-to-many.
type M2MRelation struct {
	// Table — join-таблица (room_hazard_factors).
	Table string
	// OwnerColumn — колонка FK на сущность-владельца (room_id).
	OwnerColumn string
	// ValueColumn — колонка FK на связанную запись (hazard_factor_id).
	ValueColumn string
}

// entities - реестр, наполняется проектом через Register (при старте, до middleware).
var entities []*Entity

// Register заменяет реестр сущностей (вызывается при старте проекта один раз).
func Register(list []*Entity) {
	entities = list
}

// GetEntity возвращает сущность реестра по URL-ключу.
func GetEntity(key string) *Entity {
	for _, e := range entities {
		if e.Key == key {
			return e
		}
	}
	return nil
}

// Entities возвращает текущий реестр сущностей.
func Entities() []*Entity {
	return entities
}
