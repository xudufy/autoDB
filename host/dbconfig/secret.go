package dbconfig

import "database/sql"

const (
	DBRoot = "@tcp(localhost:3306)/"
	DBACredential = "autodb:S20-CS542"
	DBRootURL = DBACredential + DBRoot + "?multiStatements=true"
	DBHostURL = DBACredential + DBRoot + "autodb?multiStatements=true"
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

	dbConnMap = make(map[string]*sql.DB)
	reservedWordSet = nil
}
