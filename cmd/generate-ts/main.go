// Команда генерации TS-классов из Go-моделей map (В1, TASKS.md).
// Команда генерации TS-классов из Go-моделей сервиса.
//
// Читает <service>/server/models (пакет models), строит Schemas (пакет codegen)
// и генерирует «оболочки» классов в <out>: поля (по json-тегам и типам Go),
// конструктор BuildClass и GetClassName.
//
// Ручные классы в src/classes/ остаются источником методов; сгенерированные
// файлы — канонический скелет полей и детектор рассинхрона (CI: make check-ts).
//
// Запуск (из каталога сервиса, где лежит models/, т.е. <service>/server):
//
//	go run github.com/pro-assistance-dev/sprob/cmd/generate-ts \
//	  -service incident [-models .] [-out ../client/src/classes/generated]
//
// Пакет codegen общий для всех сервисов экосистемы — копии не нужны.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/pro-assistance-dev/sprob/codegen"
	"github.com/pro-assistance-dev/sprob/config"
)

// headerTitle — заголовок сгенерированного файла (имя сервиса подставляется в main).
var headerTitle = `// ⚠️ GENERATED (make generate-ts) — не редактировать вручную.
// Источник: %s/server/models (Go AST → TS; генератор — sprob/codegen). Сверка в CI: make check-ts.
`

// eslintHeader — файлы с полями any (внешние типы: orb.Geometry и пр.)
// полностью отключают линтер, как принято для сгенерированного кода.
const eslintDisable = "/* eslint-disable */\n"

func main() {
	outDir := flag.String("out", "../client/src/classes/generated", "каталог для сгенерированных TS-классов")
	service := flag.String("service", "", "имя сервиса (для заголовка файлов), напр. map")
	models := flag.String("models", "", "путь к каталогу моделей (по умолчанию models относительно cwd)")
	flag.Parse()

	if *service == "" {
		fatal("не задан -service (имя сервиса для заголовка, напр. -service map)")
	}

	p := codegen.NewProject(&config.Project{ModelsPath: *models})
	schemas := sortedSchemas(p.Schemas)

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fatal("mkdir %s: %v", *outDir, err)
	}

	generated := make([]string, 0, len(schemas))
	skipped := make([]string, 0)
	for _, s := range schemas {
		if s.NameTable == "" {
			skipped = append(skipped, s.NamePascal)
			continue // не таблица (обёртки *WithCount, TokensWithUser и т.п.)
		}
		file := filepath.Join(*outDir, s.NamePascal+".ts")
		if err := os.WriteFile(file, []byte(renderClass(*service, s)), 0o644); err != nil {
			fatal("write %s: %v", file, err)
		}
		generated = append(generated, s.NamePascal)
	}

	// index.ts — реестр сгенерированных классов.
	header := headerFor(*service)
	index := header + "\n"
	for _, name := range generated {
		index += fmt.Sprintf("import %s from './%s';\n", name, name)
	}
	index += "\nconst GeneratedClasses = {\n"
	for _, name := range generated {
		index += fmt.Sprintf("  %s,\n", name)
	}
	index += "};\n\nexport { GeneratedClasses };\n"
	if err := os.WriteFile(filepath.Join(*outDir, "index.ts"), []byte(index), 0o644); err != nil {
		fatal("write index.ts: %v", err)
	}

	fmt.Printf("✅ Сгенерировано %d TS-классов → %s\n", len(generated), *outDir)
	if len(skipped) > 0 {
		fmt.Printf("   Пропущено (не таблицы): %s\n", strings.Join(skipped, ", "))
	}

	// Отчёт о рассинхроне с ручными классами src/classes/.
	runtimeDir := filepath.Clean(filepath.Join(*outDir, ".."))
	printDrift(generated, runtimeDir)
}

func sortedSchemas(schemas codegen.Schemas) []*codegen.Schema {
	keys := make([]string, 0, len(schemas))
	for k := range schemas {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]*codegen.Schema, 0, len(keys))
	for _, k := range keys {
		out = append(out, schemas[k])
	}
	return out
}

// printDrift — сравнивает сгенерированные классы с ручными в src/classes/.
func printDrift(generated []string, runtimeDir string) {
	gSet := map[string]bool{}
	for _, g := range generated {
		gSet[g] = true
	}
	entries, err := os.ReadDir(runtimeDir)
	if err != nil {
		return
	}
	rSet := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".ts" || e.Name() == "index.ts" {
			continue
		}
		rSet[strings.TrimSuffix(e.Name(), ".ts")] = true
	}

	missing := make([]string, 0) // есть в Go, нет ручного класса
	for _, g := range generated {
		if !rSet[g] {
			missing = append(missing, g)
		}
	}
	extra := make([]string, 0) // ручной класс без Go-модели
	for r := range rSet {
		if !gSet[r] {
			extra = append(extra, r)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)
	if len(missing) > 0 {
		fmt.Printf("⚠️  Go-модели без ручного TS-класса: %s\n", strings.Join(missing, ", "))
	}
	if len(extra) > 0 {
		fmt.Printf("ℹ️  Ручные TS-классы без Go-модели: %s\n", strings.Join(extra, ", "))
	}
}

// headerFor — заголовок сгенерированного файла для конкретного сервиса.
func headerFor(service string) string {
	return fmt.Sprintf(headerTitle, service)
}

// renderClass — полный текст TS-класса-оболочки.
func renderClass(service string, s *codegen.Schema) string {
	header := headerFor(service)
	var b strings.Builder

	// Импорты (связи; сам класс не импортируется). valueUse — используется ли
	// класс как значение (декоратор GetClassConstructor / `= new X()`);
	// иначе это только тип → import type (consistent-type-imports).
	imports := map[string]bool{}
	valueUse := map[string]bool{}
	for _, f := range s.Order {
		if f.SkipJSON || f.Schema == nil {
			continue
		}
		name := f.Schema.NamePascal
		if name == s.NamePascal {
			continue
		}
		imports[name] = true
		if f.IsSlice || !f.OmitEmpty {
			valueUse[name] = true
		}
	}
	names := make([]string, 0, len(imports))
	for n := range imports {
		names = append(names, n)
	}
	sort.Strings(names)

	b.WriteString(header)
	if len(names) > 0 {
		b.WriteString("\n")
		for _, n := range names {
			if valueUse[n] {
				fmt.Fprintf(&b, "import %s from './%s';\n", n, n)
			} else {
				fmt.Fprintf(&b, "import type %s from './%s';\n", n, n)
			}
		}
	}

	fmt.Fprintf(&b, "\nexport default class %s {\n", s.NamePascal)
	usesAny := false
	for _, f := range s.Order {
		if f.SkipJSON {
			continue
		}
		if c := strings.TrimSpace(f.Comment); c != "" {
			fmt.Fprintf(&b, "  // %s\n", c)
		}
		line, usedAny := renderField(s, f)
		usesAny = usesAny || usedAny
		fmt.Fprintf(&b, "  %s\n", line)
	}

	b.WriteString("\n  constructor(i?: " + s.NamePascal + ") {\n")
	b.WriteString("    PF.H.Classes.BuildClass(this, i);\n")
	b.WriteString("  }\n")
	b.WriteString("\n  static GetClassName(): string {\n")
	fmt.Fprintf(&b, "    return '%s';\n", strcase.ToLowerCamel(s.NamePascal))
	b.WriteString("  }\n")
	b.WriteString("}\n")

	content := b.String()
	if usesAny {
		content = eslintDisable + strings.TrimPrefix(content, header)
	}
	return content
}

// renderField — объявление одного поля в TS. Второй результат — использован ли any.
func renderField(s *codegen.Schema, f *codegen.Field) (string, bool) {
	name := f.NameJSON
	switch {
	case f.IsSlice:
		elem := "any"
		usedAny := false
		if f.Schema != nil {
			elem = f.Schema.NamePascal
		} else if f.ElemType != "" {
			elem = tsScalar(elemTSName(f.ElemType))
			usedAny = elem == "any"
		}
		line := fmt.Sprintf("%s: %s[] = [];", name, elem)
		if f.Schema != nil {
			return fmt.Sprintf("@PF.H.Classes.GetClassConstructor(%s)\n  %s", elem, line), false
		}
		return line, usedAny
	case f.IsUUID, f.IsTime:
		return fmt.Sprintf("%s?: string;", name), false
	case f.Schema != nil:
		// Связь: *X / X (belongs-to, has-one, m2m)
		x := f.Schema.NamePascal
		if f.OmitEmpty {
			return fmt.Sprintf("%s?: %s;", name, x), false
		}
		return fmt.Sprintf("%s = new %s();", name, x), false
	case isGoString(f.TypeGo):
		return fmt.Sprintf("%s = '';", name), false
	case isGoBool(f.TypeGo):
		return fmt.Sprintf("%s = false;", name), false
	case isGoNumber(f.TypeGo):
		return fmt.Sprintf("%s = 0;", name), false
	default:
		// orb.Geometry, baseModels.Human/FileInfo, token.Details и пр.
		return fmt.Sprintf("%s?: any;", name), true
	}
}

// elemTSName — имя типа элемента слайса для TS (без пакета).
func elemTSName(goName string) string {
	if i := strings.LastIndex(goName, "."); i >= 0 {
		return goName[i+1:]
	}
	return goName
}

func isGoString(t string) bool { return t == "string" }
func isGoBool(t string) bool   { return t == "bool" }
func isGoNumber(t string) bool { return numberRe.MatchString(t) }

var numberRe = regexp.MustCompile(`^(u?int(8|16|32|64)?|float(32|64)?)$`)

func tsScalar(goName string) string {
	switch {
	case isGoString(goName):
		return "string"
	case isGoBool(goName):
		return "boolean"
	case isGoNumber(goName):
		return "number"
	default:
		return "any"
	}
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "generate-ts: "+format+"\n", args...)
	os.Exit(1)
}
