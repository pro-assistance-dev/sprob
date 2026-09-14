package codegen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/iancoleman/strcase"
)

// Тесты общего пакета кодогенерации (вынесен из rdkb/map+incident в sprob).
// Проверяем парсинг реального пакета models на временном модуле —
// без сети/БД и без зависимости от конкретного сервиса.

// writeModule создаёт во временном каталоге минимальный Go-модуль с models/.
func writeModule(t *testing.T, modelsSrc string) string {
	t.Helper()
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "go.mod"), "module example\n\ngo 1.21\n")
	modelsDir := filepath.Join(dir, "models")
	if err := os.MkdirAll(modelsDir, 0o755); err != nil {
		t.Fatalf("mkdir models: %v", err)
	}
	writeFile(t, filepath.Join(modelsDir, "Book.go"), modelsSrc)
	return dir
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

const bookModel = `package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Book struct {
	bun.BaseModel ` + "`bun:\"books,alias:books\"`" + `

	Title     string        ` + "`json:\"title\" bun:\"title\"`" + `
	Pages     int           ` + "`json:\"pages\" bun:\"pages\"`" + `
	InStock   bool          ` + "`json:\"inStock\" bun:\"in_stock\"`" + `
	AuthorID  uuid.NullUUID ` + "`json:\"authorId\" bun:\"author_id,type:uuid\"`" + `
}

type Books []*Book

type BooksWithCount struct {
	Books Books ` + "`json:\"items\"`" + `
	Count int   ` + "`json:\"count\"`" + `
}
`

func buildSchemas(t *testing.T, dir string) Schemas {
	t.Helper()
	p := &Project{ModelsPath: dir}
	p.InitSchemas()
	if len(p.Schemas) == 0 {
		t.Fatal("схемы не построены — парсинг models не сработал")
	}
	return p.Schemas
}

func TestNewProjectParsesModels(t *testing.T) {
	dir := writeModule(t, bookModel)
	schemas := buildSchemas(t, dir)

	key := strcase.ToLowerCamel("Book")
	book, ok := schemas[key]
	if !ok {
		t.Fatalf("схема %q не найдена среди %d схем", key, len(schemas))
	}
	if book.NameTable != "books" {
		t.Errorf("имя таблицы %q, ожидалось books (множественное число)", book.NameTable)
	}
	if book.NamePascal != "Book" {
		t.Errorf("NamePascal = %q, ожидалось Book", book.NamePascal)
	}
}

func TestProjectModelsPathIsRespected(t *testing.T) {
	dir := writeModule(t, bookModel)

	// ModelsPath должен указывать на конкретный каталог: раньше параметр
	// игнорировался и всегда обходился ".". Передаём заведомо пустой каталог
	// и ожидаем, что чужие модели туда не попадут.
	empty := t.TempDir()
	p := &Project{ModelsPath: empty}
	p.InitSchemas()

	if _, ok := p.Schemas[strcase.ToLowerCamel("Book")]; ok {
		t.Error("Book попал в схемы из другого каталога — ModelsPath не учитывается")
	}
	_ = dir
}

// TestNewProjectNilConfig — регрессия: NewProject(nil) паниковал
// (разыменование config без проверки). Библиотеку вызывают из тестов/CLI,
// где конфига может не быть — конструктор обязан это выдерживать.
func TestNewProjectNilConfig(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("NewProject(nil) паникует: %v", r)
		}
	}()
	p := NewProject(nil)
	if p == nil {
		t.Fatal("NewProject(nil) вернул nil")
	}
	if p.ModelsPath == "" {
		t.Error("ModelsPath по умолчанию не должен быть пустым")
	}
}

func TestSchemasSkipNonTables(t *testing.T) {
	dir := writeModule(t, bookModel)
	schemas := buildSchemas(t, dir)

	// BooksWithCount — обёртка без bun-таблицы, у неё NameTable пуст.
	wrapper, ok := schemas[strcase.ToLowerCamel("BooksWithCount")]
	if !ok {
		t.Skip("обёртка *WithCount не построена — поведение парсера отличается")
	}
	if wrapper.NameTable != "" {
		t.Errorf("у обёртки BooksWithCount не должно быть таблицы, получено %q", wrapper.NameTable)
	}
}

func TestBookFieldsParsed(t *testing.T) {
	dir := writeModule(t, bookModel)
	schemas := buildSchemas(t, dir)

	book := schemas[strcase.ToLowerCamel("Book")]
	if book == nil {
		t.Fatal("схема Book не найдена")
	}

	// json-имена должны сохраняться как есть (в camelCase, как в модели).
	wantJSON := map[string]bool{"title": false, "pages": false, "inStock": false, "authorId": false}
	for _, f := range book.Order {
		if _, ok := wantJSON[f.NameJSON]; ok {
			wantJSON[f.NameJSON] = true
		}
	}
	for name, found := range wantJSON {
		if !found {
			t.Errorf("поле %q не найдено в схеме Book", name)
		}
	}
}

func TestOrderPreservesDeclaration(t *testing.T) {
	dir := writeModule(t, bookModel)
	schemas := buildSchemas(t, dir)

	book := schemas[strcase.ToLowerCamel("Book")]
	if len(book.Order) == 0 {
		t.Fatal("порядок полей пуст")
	}
	first := book.Order[0].NameJSON
	if first != "title" {
		t.Errorf("первое поле = %q, ожидалось title (порядок объявления)", first)
	}
}

func TestInitFieldsLinksToSchemas(t *testing.T) {
	dir := writeModule(t, `package models

type Genre struct {
	Name string `+"`json:\"name\" bun:\"name\"`"+`
}

type Book struct {
	Title  string `+"`json:\"title\" bun:\"title\"`"+`
	Genre  *Genre `+"`json:\"genre\" bun:\"rel:belongs-to\"`"+`
}

type Books []*Book
type Genres []*Genre
`)
	schemas := buildSchemas(t, dir)

	book := schemas[strcase.ToLowerCamel("Book")]
	if book == nil {
		t.Fatal("схема Book не найдена")
	}

	var linked bool
	for _, f := range book.Order {
		if f.NameJSON == "genre" {
			if f.Schema == nil {
				t.Error("связь genre не получила схему Genre (InitFieldsLinksToSchemas)")
			} else if f.Schema.NamePascal != "Genre" {
				t.Errorf("связь genre указывает на %q, ожидалось Genre", f.Schema.NamePascal)
			}
			linked = true
		}
	}
	if !linked {
		t.Error("поле genre не найдено в схеме Book")
	}
}

func TestInitSchemasIsIdempotent(t *testing.T) {
	dir := writeModule(t, bookModel)
	p := NewProject(nil)
	p.ModelsPath = dir
	p.InitSchemas()
	n := len(p.Schemas)

	// Повторный вызов не должен переразбирать/дублировать схемы.
	p.InitSchemas()
	if len(p.Schemas) != n {
		t.Errorf("повторный InitSchemas изменил число схем: %d → %d", n, len(p.Schemas))
	}
}

func TestSchemasLibFilled(t *testing.T) {
	dir := writeModule(t, bookModel)
	_ = buildSchemas(t, dir)

	if _, ok := SchemasLib[strcase.ToLowerCamel("Book")]; !ok {
		t.Error("SchemasLib не заполнен после InitSchemas")
	}
}
