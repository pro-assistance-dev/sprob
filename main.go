package main

import (
	"fmt"

	"github.com/pro-assistance/pro-assister/helpers/project"
)

func main() {
	t := project.NewProject()
	t.InitSchemas()
	for _, v := range t.Schemas {
		for i, t := range v {
			fmt.Println(i, t)
		}
	}
}
