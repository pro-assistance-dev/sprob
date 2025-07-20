package project

import (
	"fmt"
	"go/parser"
	"go/token"
	"log"
	"testing"

	"github.com/iancoleman/strcase"
	"github.com/stretchr/testify/assert"
)

var p = Project{Schemas: make(SchemasMap, 0)}

func ProjectTestSetup() {
	modelsPackage, err := parser.ParseDir(token.NewFileSet(), "./mocks/", nil, parser.AllErrors)
	if err != nil {
		log.Fatal(err)
	}
	structs := p.getStructsOfProject(modelsPackage)

	for s := range structs {
		schema := newSchema(s, structs[s])
		key := strcase.ToLowerCamel(s.Name.String())
		p.Schemas[key] = &schema
		SchemasSlice = append(SchemasSlice, &schema)
	}
	p.Schemas.InitFieldsLinksToSchemas()
	SchemasLib = p.Schemas
}

func TestProject(t *testing.T) {
	ProjectTestSetup()

	fmt.Println(SchemasLib)
	t.Run("SchemasLen", func(t *testing.T) {
		assert.Equal(t, 3, len(SchemasLib), "When 3 struct are defined, len schemas should be 3")
	})

	t.Run("GetSchemas", func(t *testing.T) {
		assert.NotNil(t, p.Schemas.GetSchema("contact"), "Find existing struct")
		assert.Nil(t, p.Schemas.GetSchema("Contact"), "When find struct with PascalCase, return nil")
		assert.Nil(t, p.Schemas.GetSchema("NotExistst"), "When struct not exist, return nil")
	})

	// schema := p.Schemas.GetSchema("contact")

	t.Run("SchemaTest", func(t *testing.T) {
		// t.Run("AllNamesCorrects", func(t *testing.T) {
		// 	assert.Equal(t, "Contact", schema.NamePascal, "NamePascal")
		// 	assert.Equal(t, "name", schema.SortColumn, "SortColumn")
		// 	assert.Equal(t, "contacts", schema.NameTable, "NameTable")
		// })
		// fmt.Println(p.Schemas)
		// t.Run("HaveCorrectFieldsLen", func(t *testing.T) {
		// 	assert.Equal(t, 3, len(schema.Fields), "When 3 fields defined, len fields should be 3")
		// })

		// fields := []string{"id", "name", "emails"}
		// t.Run("HaveCorrectFields", func(t *testing.T) {
		// 	for _, f := range fields {
		// 		t.Run(f, func(t *testing.T) {
		// 			assert.Equal(t, schema.GetField(f).NameCamel, f, "Field "+f)
		// 		})
		// 	}
		// })

		// t.Run("ConcatTableColCorrectly", func(t *testing.T) {
		// 	for _, f := range fields {
		// 		t.Run(f, func(t *testing.T) {
		// 			assert.Equal(t, schema.ConcatTableCol(f), "contacts"+"."+f, "Conctat field "+f)
		// 		})
		// 	}
		// })
		//
		// t.Run("ConcatTableCols", func(t *testing.T) {
		// 	cols := schema.ConcatTableCols()
		// 	assert.Equal(t, len(cols), 2, "Len eq 2")
		// })
	})
}
