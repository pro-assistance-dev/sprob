package targets

import (
	"context"
	"strings"

	"github.com/pro-assistance-dev/sprob/modules/survey/models"
)

func (s *Service) ListByTarget(c context.Context, targetType, targetID string) (models.Publications, error) {
	return s.r.ListByTarget(c, targetType, targetID)
}

// EnsureForm — создать (или вернуть существующую) форму+публикацию для сущности.
func (s *Service) EnsureForm(c context.Context, spec FormSpec) (*models.Publication, error) {
	normalizeSpec(&spec)
	return s.r.EnsureForm(c, spec)
}

// normalizeSpec — значения по умолчанию + slug из кода полей.
func normalizeSpec(spec *FormSpec) {
	if strings.TrimSpace(spec.SectionName) == "" {
		spec.SectionName = "Анкета"
	}
	if spec.Title == "" {
		spec.Title = spec.EntityName
	}
	if spec.FormName == "" {
		spec.FormName = spec.Title
	}
	for i := range spec.Fields {
		if strings.TrimSpace(spec.Fields[i].ValueType) == "" {
			spec.Fields[i].ValueType = "string"
		}
		if spec.Fields[i].Code == "" {
			spec.Fields[i].Code = slugify(spec.Fields[i].Name)
		}
	}
}

// slugify — латиница-код из имени (кириллицу не транслитерируем буквально —
// для кодов важна стабильность, поэтому используем безопасный короткий вид).
func slugify(name string) string {
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '_' || r == '-':
			b.WriteRune('_')
		}
	}
	out := b.String()
	if out == "" {
		return "field"
	}
	return out
}
