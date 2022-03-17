package sqlHelper

import (
	"fmt"
)

type SQLHelper struct {
}

func NewSQLHelper() *SQLHelper {
	return &SQLHelper{}
}

func (i *SQLHelper) WhereLikeWithLowerTranslit(col string, search string) string {
	return fmt.Sprintf("WHERE lower(regexp_replace(%s, '[^а-яА-Яa-zA-Z0-9 ]', '', 'g')) LIKE lower(%s)", col, "'%"+search+"%'")
}
