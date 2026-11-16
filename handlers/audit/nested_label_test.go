package audit

import "testing"

// ⚠️ Вложенные объекты в дельте (15.11). В теле запроса клиент присылает
// связи ЦЕЛИКОМ (`worker: {human: {...}, contact: {...}}`), и в «стало»
// попадал сырой JSON — «Contact: — → {"description":"","emails":[…]}».
// Владелец: «непонятно, что произошло». Теперь — человеческая подпись.
func TestNestedLabel(t *testing.T) {
	cases := []struct {
		name string
		in map[string]interface{}
		want string
	}{
		{
			name: "ФИО из human",
			in: map[string]interface{}{
				"human": map[string]interface{}{
					"surname": "Прикот", "name": "Галина", "patronymic": "Михайловна",
				},
			},
			want: "Прикот Галина Михайловна",
		},
		{
			name: "ФИО плоско",
			in: map[string]interface{}{"surname": "Иванов", "name": "Пётр"},
			want: "Иванов Пётр",
		},
		{
			name: "название связи",
			in: map[string]interface{}{"name": "Прививка от кори"},
			want: "Прививка от кори",
		},
		{
			name: "код, если нет названия",
			in: map[string]interface{}{"code": "C0123"},
			want: "C0123",
		},
		{
			name: "пустой объект — без подписи",
			in: map[string]interface{}{"description": "", "emails": []interface{}{}},
			want: "",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := nestedLabel(c.in); got != c.want {
				t.Errorf("nestedLabel() = %q, want %q", got, c.want)
			}
		})
	}
}

// humanValue не должен печатать JSON вложенного объекта.
func TestHumanValueNestedNotJSON(t *testing.T) {
	got := humanValue("contact", map[string]interface{}{
		"description": "", "emails": []interface{}{map[string]interface{}{"number": "1"}},
	})
	if got != "ссылка" {
		t.Errorf("humanValue(вложенный объект) = %q, хотим «ссылка», а не JSON", got)
	}
	if got == "{" || len(got) > 12 {
		t.Errorf("humanValue вернул JSON: %q", got)
	}
}
