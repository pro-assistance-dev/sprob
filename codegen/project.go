package codegen

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/pro-assistance-dev/sprob/config"
)

const defaultModelDir = "models"

type Project struct {
	Schemas    Schemas `json:"schemas"`
	ModelsPath string
}

// NewProject готовит Project и сразу разбирает каталог моделей.
// config может быть nil — тогда ModelsPath берётся по умолчанию ("models");
// это позволяет использовать пакет без конфига (тесты, CLI-флаги).
func NewProject(cfg *config.Project) *Project {
	modelsPath := defaultModelDir
	if cfg != nil && cfg.ModelsPath != "" {
		modelsPath = cfg.ModelsPath
	}
	// ModelsPath == "" означает «искать в текущем каталоге» (историческое
	// поведение сервисных копий генератора: обход "."). Иначе — конкретный путь.
	p := &Project{ModelsPath: modelsPath}
	p.InitSchemas()
	return p
}

var SchemasLib = Schemas{}

// findAllModelsPackages обходит деревья каталогов, начиная с root, собирая
// каталоги-кандидаты для парсинга (пропускает те, что вне модели).
// root по умолчанию ""." — как было у сервисных копий генератора.
func findAllModelsPackages(root string) []string {
	if root == "" {
		root = "."
	}
	paths := make([]string, 0)
	err := filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() || strings.Contains(path, "static") {
				return nil
			}
			paths = append(paths, path)
			return nil
		})
	if err != nil {
		log.Println(err)
	}
	return paths
}

func (i *Project) InitSchemas() {
	if len(i.Schemas) > 0 {
		return
	}
	// ModelsPath позволяет запускать генератор из любого каталога
	// (-models <service>/server); по умолчанию — текущий (".").
	paths := findAllModelsPackages(i.ModelsPath)
	i.Schemas = make(Schemas, 0)
	for _, path := range paths {
		modelsPackage, err := parser.ParseDir(token.NewFileSet(), path, nil, parser.AllErrors)
		if err != nil {
			log.Fatal(err)
		}
		structs, typeLookup := i.getStructsOfProject(modelsPackage)

		for s := range structs {
			schema := newSchema(s, structs[s], typeLookup)
			key := strcase.ToLowerCamel(s.Name.String())
			i.Schemas[key] = &schema
		}
	}

	i.Schemas.InitFieldsLinksToSchemas()
	SchemasLib = i.Schemas
}

func (i *Project) getStructsOfProject(modelsPackage map[string]*ast.Package) (map[*ast.TypeSpec][]*ast.Field, map[string]*ast.TypeSpec) { //nolint:all
	structs := map[*ast.TypeSpec][]*ast.Field{}
	typeLookup := map[string]*ast.TypeSpec{}

	pack := modelsPackage["models"]
	if pack == nil {
		pack = modelsPackage["mocks"]
	}
	if pack == nil {
		return nil, nil
	}

	for _, file := range pack.Files {
		for _, node := range file.Decls {
			genDecl, ok := node.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				// Все типы пакета (структуры и алиасы слайсов) — для разрешения связей.
				typeLookup[typeSpec.Name.Name] = typeSpec
				switch tt := typeSpec.Type.(type) {
				case *ast.StructType:
					structs[typeSpec] = tt.Fields.List
				}
			}
		}
	}

	return structs, typeLookup
}
