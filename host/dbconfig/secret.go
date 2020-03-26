package dbconfig

import "database/sql"

const (
	//DBBaseURL todo
	DBRootURL = "autodb:S20-CS542@tcp(localhost:3306)/"
	DBHostURL = "autodb:S20-CS542@tcp(localhost:3306)/autodb"
)

var HostDB *sql.DB
var RootDB *sql.DB

func Init() {
	tHostDB, err := sql.Open("mysql", DBHostURL)
	if err != nil {
		panic("autodb DB connection failed.")
	}
	HostDB = tHostDB
	tRootDB, err := sql.Open("mysql", DBRootURL)
	if err != nil {
		panic("root DB connection failed.")
	}
	RootDB = tRootDB
}
