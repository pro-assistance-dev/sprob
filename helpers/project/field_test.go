package project

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSchemaField(t *testing.T) {
	field := NewSchemaField("1", "2", "3", "4")

	t.Run("", func(t *testing.T) {
		t.Run("AlllNamesCorrect", func(t *testing.T) {
			assert.Equal(t, "NamePascal", field.NamePascal, "1")
			assert.Equal(t, "NameCol", field.NameCol, "2")
			assert.Equal(t, "NameCamel", field.NameCamel, "1")
			assert.Equal(t, "NameRus", field.NameRus, "3")
		})

		t.Run("TypeCorrect", func(t *testing.T) {
			assert.Equal(t, "Type", field.Type, "4")
		})
	})
}
