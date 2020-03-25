package main

import (
	"autodb/host/dbconfig"
	"autodb/host/globalsession"
	"autodb/host/handler"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
)

func main() {

	exePath := os.Args[0]
	err := os.Chdir(filepath.Dir(exePath))

	globalsession.Init()

	handler.InitAllHTTPHandlers()

	sqlTest()

	port := 23456
	portS := fmt.Sprintf(":%d", port)
	fmt.Printf("listening on %s", portS)
	err = http.ListenAndServe(portS, nil)
	if err != nil {
		fmt.Printf("%v", err)
	}
}

func sqlTest() {
	_, err := sql.Open("mysql", dbconfig.DBBaseURL)
	if err != nil {
		fmt.Println(err)
		return
	}
	dbHost, err := sql.Open("mysql", dbconfig.DBHostURL)
	if err != nil {
		fmt.Println(err)
		return
	}
	tables, err := dbHost.Query("show tables;")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tables.Close()

	for tables.Next() {
		var tablename string
		err := tables.Scan(&tablename)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(tablename)
	}
}
