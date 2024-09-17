package project

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/pro-assistance/pro-assister/config"
)

const defaultModelDir = "models"

type Project struct {
	Schemas    Schemas `json:"schemas"`
	ModelsPath string
}

func NewProject(config *config.Project) *Project {
	modelsPath := config.ModelsPath
	if modelsPath == "" {
		modelsPath = defaultModelDir
	}
	p := &Project{ModelsPath: modelsPath}
	p.InitSchemas()
	return p
}

var SchemasLib = Schemas{}

func findAllModelsPackages() []string {
	paths := make([]string, 0)
	err := filepath.Walk(".",
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
	paths := findAllModelsPackages()

	i.Schemas = make(Schemas, 0)
	for _, path := range paths {
		modelsPackage, err := parser.ParseDir(token.NewFileSet(), path, nil, parser.AllErrors)
		if err != nil {
			log.Fatal(err)
		}

		structs := i.getStructsOfProject(modelsPackage)

		for s := range structs {
			schema := newSchema(s, structs[s])
			key := strcase.ToLowerCamel(s.Name.String())
			// fmt.Println(key)
			i.Schemas[key] = &schema
		}
	}

	i.Schemas.InitFieldsLinksToSchemas()
	SchemasLib = i.Schemas
}

func (i *Project) getStructsOfProject(modelsPackage map[string]*ast.Package) map[*ast.TypeSpec][]*ast.Field { //nolint:all
	structs := map[*ast.TypeSpec][]*ast.Field{}

	pack := modelsPackage["models"]
	if pack == nil {
		pack = modelsPackage["mocks"]
	}
	if pack == nil {
		return nil
	}

	for _, file := range pack.Files {
		for _, node := range file.Decls {
			switch node.(type) {
			case *ast.GenDecl:
				genDecl := node.(*ast.GenDecl)
				for _, spec := range genDecl.Specs {
					switch spec.(type) {
					case *ast.TypeSpec:
						typeSpec := spec.(*ast.TypeSpec)
						switch typeSpec.Type.(type) {
						case *ast.StructType:
							structType := typeSpec.Type.(*ast.StructType)
							structs[typeSpec] = structType.Fields.List
						}
					}
				}
			}
		}
	}

	return structs
}
