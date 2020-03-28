package main

import (
	"autodb/host/dbconfig"
	"autodb/host/globalsession"
	"autodb/host/handler"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
)

func main() {

	exePath := os.Args[0]
	err := os.Chdir(filepath.Dir(exePath))

	dbconfig.Init()

	globalsession.Init()

	handler.InitAllHTTPHandlers()

	sqlTest()

	port := 23456
	portS := fmt.Sprintf(":%d", port)
	fmt.Printf("listening on %s\n", portS)
	err = http.ListenAndServe(portS, nil)
	if err != nil {
		fmt.Printf("%v", err)
	}
}

func sqlTest() {

	tables, err := dbconfig.HostDB.Query("select * from nullTest")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tables.Close()

	//for tables.Next() {
	//	var tablename string
	//	err := tables.Scan(&tablename)
	//	if err != nil {
	//		fmt.Println(err)
	//	}
	//	fmt.Println(tablename)
	//}
	js, err:=dbconfig.ParseRowsToJSON(tables)
	if err!=nil {
		fmt.Println(err)
		return
	}
	fmt.Println(js)
}
