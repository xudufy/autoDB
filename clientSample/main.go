package main

import (
	"autodb/host/globalsession"
	"autodb/host/handler"
	"fmt"
)

func main() {
	fmt.Println("Hello, world")
	handler.InitAllHTTPHandlers()
	globalsession.Init()
	//http.ListenAndServe()
}
