package project

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"

	"github.com/pro-assistance/pro-assister/config"
)

const defaultModelDir = "models"

type Project struct {
	Schemas
	ModelsPath string
}

func NewProject(config *config.Config) *Project {
	modelsPath := config.ModelsPath
	if modelsPath == "" {
		modelsPath = defaultModelDir
	}
	return &Project{ModelsPath: modelsPath}
}

var SchemasLib = Schemas{}

func (i *Project) InitSchemas() {
	modelsPackage, err := parser.ParseDir(token.NewFileSet(), i.ModelsPath, nil, parser.AllErrors)
	if err != nil {
		log.Fatal(err)
	}
	structs := i.getStructsOfProject(modelsPackage)
	i.Schemas = make(Schemas, 0)
	for s := range structs {
		fmt.Printf("%+v\n", s.Type)
		// schema := newSchema(s, structs[s])
		// i.Schemas[ToLowerCamel(s.Name.String())] = schema
	}
	SchemasLib = i.Schemas
}

func (i *Project) getStructsOfProject(modelsPackage map[string]*ast.Package) map[*ast.TypeSpec][]*ast.Field { //nolint:all
	structs := map[*ast.TypeSpec][]*ast.Field{}
	for _, file := range modelsPackage["models"].Files {
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
