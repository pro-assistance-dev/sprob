package targets

import (
	"testing"
)

func TestNormalizeSpec(t *testing.T) {
	spec := FormSpec{
		TargetType: "event",
		TargetID:   "d0000000-0000-4000-8000-000000000001",
		EntityName: "Конференция",
		Fields: []FieldSpec{
			{Name: "ФИО"},                            // без типа/кода
			{Name: "Email", ValueType: "string"},     // тип есть
			{Name: "Откуда узнали?", Code: "source"}, // код задан
		},
	}
	normalizeSpec(&spec)
	if spec.SectionName != "Анкета" {
		t.Errorf("SectionName = %q", spec.SectionName)
	}
	if spec.Title != "Конференция" || spec.FormName != "Конференция" {
		t.Errorf("Title/FormName = %q/%q", spec.Title, spec.FormName)
	}
	if spec.Fields[0].ValueType != "string" {
		t.Errorf("тип по умолчанию не подставлен: %q", spec.Fields[0].ValueType)
	}
	if spec.Fields[0].Code != "field" { // кириллица → пустой код → "field"
		t.Errorf("Code кириллицы = %q", spec.Fields[0].Code)
	}
	if spec.Fields[1].Code != "Email" {
		t.Errorf("Code = %q", spec.Fields[1].Code)
	}
	if spec.Fields[2].Code != "source" {
		t.Errorf("Code перезаписан: %q", spec.Fields[2].Code)
	}
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Email":      "Email",
		"first name": "first_name",
		"a-b":        "a_b",
		"":           "field",
		"ФИО":        "field", // транслитерацию не делаем — безопасный короткий код
	}
	for in, want := range cases {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBuildForm(t *testing.T) {
	form := buildForm(FormSpec{
		FormName:    "Регистрация",
		SectionName: "Анкета",
		Fields: []FieldSpec{
			{Name: "ФИО", Code: "fio", Required: true, ValueType: "string"},
			{Name: "Тип", Code: "type", ValueType: "radio", Variants: []string{"A", "B"}},
		},
	}, nil)
	if len(form.FormSections) != 1 {
		t.Fatalf("секций = %d", len(form.FormSections))
	}
	fields := form.FormSections[0].Fields
	if len(fields) != 2 {
		t.Fatalf("полей = %d", len(fields))
	}
	if !fields[0].Required || fields[0].Code != "fio" {
		t.Errorf("поле 0: %+v", fields[0])
	}
	if len(fields[1].AnswerVariants) != 2 {
		t.Errorf("вариантов = %d", len(fields[1].AnswerVariants))
	}
	if fields[1].AnswerVariants[0].FieldID != fields[1].ID {
		t.Error("FieldID варианта не привязан к полю")
	}
}
