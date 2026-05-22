package main

import (
	"fmt"

	"github.com/keainya/pages/object"
	"github.com/keainya/pages/router"
)

func main() {
	if object.Database == nil {
		fmt.Println("database error")
	}
	router.InitRouter(webFS).Run()
}
