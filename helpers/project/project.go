package project

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"slices"
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

func addToPaths(paths []string, path string, info os.FileInfo, err error) ([]string, error) {
	if err != nil {
		return nil, err
	}
	if !info.IsDir() || strings.Contains(path, "static") || strings.Contains(path, "modules") || strings.Contains(path, "logs") || strings.Contains(path, ".vscode") {
		return nil, nil
	}
	paths = append(paths, path)
	return paths, nil
}

func findAllModelsPackages() []string {
	ctx := build.Default
	pkg, err := ctx.Import("github.com/pro-assistance/pro-assister/models", ".", build.FindOnly)
	if err != nil {
		panic(err)
	}
	pathsToParse := []string{".", pkg.Dir}
	paths := make([]string, 0)

	for _, p := range pathsToParse {
		err := filepath.Walk(p,
			func(path string, info os.FileInfo, err error) error {
				fmt.Println("paths", paths, p)
				pathsToAdd, err := addToPaths(paths, path, info, err)
				if err != nil {
					return err
				}

				if pathsToAdd != nil && !slices.Contains(paths, path) {
					paths = append(paths, pathsToAdd...)
				}
				return nil
			})
		if err != nil {
			log.Println(err)
		}
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
		fmt.Println("modelsPackage", modelsPackage)
		structs := i.getStructsOfProject(modelsPackage)

		for s := range structs {
			schema := newSchema(s, structs[s])
			key := strcase.ToLowerCamel(s.Name.String())
			i.Schemas[key] = &schema
		}
	}

	i.Schemas.InitFieldsLinksToSchemas()
	fmt.Println(SchemasLib)
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
