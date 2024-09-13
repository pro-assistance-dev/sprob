package project

import (
	"log"
	"testing"

	"github.com/pro-assistance/pro-assister/config"
	"github.com/stretchr/testify/assert"
)

var p *Project

func ProjectTestSetup() {
	conf, err := config.LoadTestConfig()
	if err != nil {
		log.Fatal(err)
	}
	p = NewProject(&conf.Project)
}

func TestProject(t *testing.T) {
	ProjectTestSetup()

	t.Run("SchemasLen", func(t *testing.T) {
		assert.Equal(t, 2, len(SchemasLib), "When 2 struct are defined, len schemas should be 2")
	})

	t.Run("GetSchemas", func(t *testing.T) {
		assert.NotNil(t, p.Schemas.GetSchema("contact"), "Find existing struct")
		assert.Nil(t, p.Schemas.GetSchema("Contact"), "When find struct with PascalCase, return nil")
		assert.Nil(t, p.Schemas.GetSchema("NotExistst"), "When struct not exist, return nil")
	})

	schema := p.Schemas.GetSchema("contact")

	t.Run("SchemaTest", func(t *testing.T) {
		t.Run("AllNamesCorrects", func(t *testing.T) {
			assert.Equal(t, "Contact", schema.Name, "Name")
			assert.Equal(t, "name", schema.SortColumn, "SortColumn")
			assert.Equal(t, "contacts", schema.TableName, "TableName")
		})

		t.Run("Have correct Fields len", func(t *testing.T) {
			assert.Equal(t, 3, len(schema.Fields), "When 3 fields defined, len fields should be 3")
		})
	})
}
