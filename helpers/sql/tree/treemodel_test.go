package tree

import (
	"log"
	"testing"

	"github.com/pro-assistance/pro-assister/config"
	"github.com/pro-assistance/pro-assister/helpers/project"
)

func prepare() *TreeModel {
	conf, err := config.LoadTestConfig()
	if err != nil {
		log.Fatal(err)
	}
	p := project.NewProject(conf)
	p.InitSchemas()
	return &TreeModel{}
}

func TestGetTableAndCols(t *testing.T) {
	tr := prepare()
	tr.Model = "address"
	tests := []struct {
		name  string
		input []string
		want  string
	}{
		{"TwoCols", []string{"id", "city"}, "address.id, address.city"},
		{"SkipNotExistsCol", []string{"id", "unknown"}, "address.id"},
	}
	for _, tt := range tests {
		tr.Cols = tt.input
		t.Run(tt.name, func(t *testing.T) {
			// s := tr.getTableAndCols()
			// if tt.want != s {
			// 	t.Errorf("\n got: \n %s, \n want: \n %s", s, tt.want)
			// }
		})
	}
}
