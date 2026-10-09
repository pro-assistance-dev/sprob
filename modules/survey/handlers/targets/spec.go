package targets

// FormSpec — спецификация «формы по требованию» для сущности. Модуль НЕ знает
// домен события/работника — набор полей передаёт вызывающая сторона (клиент или
// сервис), поэтому одна и та же ручка годится под регистрацию на событие,
// опросник и т.п.
type FormSpec struct {
	// TargetType/TargetID — к какой сущности привязать (portal: "event").
	TargetType string `json:"targetType" binding:"required"`
	TargetID   string `json:"targetId" binding:"required"`

	// Title — заголовок публикации (напр. «Регистрация на «Конференция»»).
	Title string `json:"title"`
	// EntityName — имя сущности (для заголовка по умолчанию).
	EntityName string `json:"entityName"`
	// FormName/SectionName — имя формы и её секции (по умолчанию от Title).
	FormName    string `json:"formName"`
	SectionName string `json:"sectionName"`
	// Slug — желаемый slug ссылки /f/<slug> (пусто — случайный).
	Slug string `json:"slug"`
	// Brandbook — бренд рантайма (rdkb/portal/pros/ferma).
	Brandbook string `json:"brandbook"`

	// Fields — поля формы по порядку.
	Fields []FieldSpec `json:"fields"`
}

// FieldSpec — одно поле «формы по требованию».
type FieldSpec struct {
	Name string `json:"name" binding:"required"`
	// Code — код поля (для параметров/интеграций); пусто — сгенерируется из имени.
	Code string `json:"code"`
	// ValueType — имя типа ответа: string|text|number|date|radio|set|boolean.
	ValueType string `json:"valueType"`
	Required  bool   `json:"required"`
	// Variants — варианты для radio/set.
	Variants []string `json:"variants"`
}
