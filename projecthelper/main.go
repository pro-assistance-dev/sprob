package projecthelper

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
)

const defaultModelDir = "models"

type ProjectHelper struct {
	Schemas
}

func NewProjectHelper() *ProjectHelper {
	return &ProjectHelper{}
}

var SchemasLib = Schemas{}

func (i *ProjectHelper) InitSchemas() {
	modelsPackage, err := parser.ParseDir(token.NewFileSet(), defaultModelDir, nil, parser.AllErrors)
	if err != nil {
		log.Fatal(err)
	}
	structs := i.getStructsOfProject(modelsPackage)
	i.Schemas = make(Schemas, 0)
	for s := range structs {
		i.Schemas[ToLowerCamel(s.Name.String())] = getSchema(s, structs[s])
	}
	SchemasLib = i.Schemas
}

func (i *ProjectHelper) getStructsOfProject(modelsPackage map[string]*ast.Package) map[*ast.TypeSpec][]*ast.Field {
	var structs = map[*ast.TypeSpec][]*ast.Field{}
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
