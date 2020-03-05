package developer

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"log"
)


const username = "dbclass"
const password = "kjfjlkfsl;fsdljfsjkfsdlkjfg09128upoiewjqdlksau8392ioewjk"
const database = "autodb"

var DB *sql.DB

func InitializeRemoteDB() bool {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", username, password, database))
	if err != nil {
		log.Fatal(err)
		return false
	} 
	err = db.Ping()
	if err != nil {
		
	}
	fmt.Println(err)
	DB = db

	return true
}


func CloseDB() bool {
	DB.Close()
	return true
}
// Query Template
// func Query(query string) sql.DB {
// 	rows, err := remDB.Query(query)


// 	if err != nil {
// 		log.Fatal(err)
// 	}

// for rows.Next() {
// 	err := rows.Scan(&id, &name)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	log.Println(id, name)
// }
// err = rows.Err()
// if err != nil {
// 	log.Fatal(err)
// }
