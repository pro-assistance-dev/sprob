package main

import (
	"fmt"

	"github.com/pro-assistance/pro-assister/socialHelper"
)

func main() {
	s := socialHelper.NewSocial()
	fmt.Println(s.GetWebFeed()[1].Link)
}
