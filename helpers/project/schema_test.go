package project

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchema(t *testing.T) {
	// ProjectTestSetup()

	t.Run("GetSchemas", func(t *testing.T) {
		assert.NotNil(t, p.Schemas.GetSchema("contact"), "Find existing struct (lowerCamel)")
		assert.NotNil(t, p.Schemas.GetSchema("Contact"), "PascalCase нормализуется в lowerCamel (Н11)")
		assert.NotNil(t, p.Schemas.GetSchema("email-message"), "kebab-case нормализуется в lowerCamel (Н11)")
		assert.Nil(t, p.Schemas.GetSchema("NotExistst"), "When struct not exist, return nil")
	})

	schema := p.Schemas.GetSchema("contact")

	t.Run("SchemaTest", func(t *testing.T) {
		t.Run("AllNamesCorrects", func(t *testing.T) {
			assert.Equal(t, "Contact", schema.NamePascal, "NamePascal")
			assert.Equal(t, "name", schema.SortColumn, "SortColumn")
			assert.Equal(t, "contacts", schema.NameTable, "NameTable")
		})
	})
}
