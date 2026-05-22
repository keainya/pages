package main

import (
	"flag"
	"fmt"

	"github.com/keainya/pages/object"
	"github.com/keainya/pages/router"
)

func main() {
	port := flag.Int("port", 8082, "HTTP 服务端口")
	flag.Parse()

	if object.Database == nil {
		fmt.Println("database error")
	}
	router.InitRouter(webFS).Run(fmt.Sprintf(":%d", *port))
}
