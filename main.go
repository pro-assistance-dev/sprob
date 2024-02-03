package main

import (
	"fmt"

	"github.com/pro-assistance/pro-assister/projecthelper"
)

func main() {
	t := projecthelper.NewProjectHelper()
	t.InitSchemas()
	for _, v := range t.Schemas {
		for i, t := range v {
			fmt.Println(i, t)
		}
	}

}
