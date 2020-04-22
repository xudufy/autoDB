package dbconfig

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

const (
	SchemaPw = "OSF3t-t2S3tG2S-4t4Et2"
)

var dbConnMap map[string]*sql.DB // init in secret.go:Init()
var connMapLock sync.RWMutex

func addConn(url string, db *sql.DB) {
	connMapLock.Lock()
	defer connMapLock.Unlock()
	if len(dbConnMap)>2000 {
		for _, conn := range dbConnMap {
			_ = conn.Close()
		}
		dbConnMap = make(map[string]*sql.DB)
	}

	_, ok := dbConnMap[url]
	if ok {
		_ = dbConnMap[url].Close()
	}
	dbConnMap[url] = db
}

func getConn(url string) (*sql.DB, error) {
	connMapLock.RLock()
	db, ok := dbConnMap[url]
	connMapLock.RUnlock()
	if !ok {
		dbc, err := sql.Open("mysql", url)
		if err!=nil {
			fmt.Println(url, err)
			return nil, err
		}
		addConn(url, dbc)
		db = dbc
	}
	return db, nil
}

func dropConn(url string) {
	connMapLock.Lock()
	defer connMapLock.Unlock()
	db, ok := dbConnMap[url]
	if ok {
		_ = db.Close()
		delete(dbConnMap, url)
	}
}

func ComposeDBUrl(username string, pw string, schemaName string) string {
	return username+":"+pw+DBRoot+schemaName
}

func GetProjectInternalConn(pid int, schemaName string) (*sql.DB, error) {
	url := composeProjectInternalDBUrl(pid, schemaName)
	return getConn(url)
}

func GetProjectPublicConn(pid int, schemaName string) (*sql.DB, error) {
	url := composeProjectPublicDBUrl(pid, schemaName)
	return getConn(url)
}

func composeProjectInternalDBUrl(pid int, schemaName string) string {
	return ComposeDBUrl(strconv.Itoa(pid)+"_internal", SchemaPw, schemaName)
}

func composeProjectPublicDBUrl(pid int, schemaName string) string {
	return ComposeDBUrl(strconv.Itoa(pid)+"_public", SchemaPw, schemaName)
}

// see in https://dev.mysql.com/doc/refman/8.0/en/grant.html#grant-quoting for the reason we need this.
func grantEscape(schemaName string) string {
	var result strings.Builder
	result.WriteRune('`')
	for _, ch:= range schemaName {
		if ch=='_' {
		result.WriteRune('\\')
		}
		result.WriteRune(ch)
	}
	result.WriteRune('`')
	return result.String()
}

func AddProjectUser(pid int, schemaName string) error {
	tx, err := RootDB.Begin()
	if err!=nil {
		return err
	}

	internalUser := `'`+ strconv.Itoa(pid) + `_internal'@'%'`
	publicUser := `'`+ strconv.Itoa(pid) + `_public'@'%'`

	var queries strings.Builder
	queries.WriteString(`DROP USER IF EXISTS `+ internalUser +`, `+ publicUser + `; ` + "\n")
	queries.WriteString(`CREATE USER `+ internalUser +` IDENTIFIED BY '`+ SchemaPw +`'; `+ "\n")
	queries.WriteString(`CREATE USER `+ publicUser + ` IDENTIFIED BY '`+SchemaPw+`'; `+ "\n")
	queries.WriteString(`GRANT ALL PRIVILEGES ON ` + grantEscape(schemaName) + `.* TO ` + internalUser + `; `+ "\n")
	queries.WriteString(`GRANT DELETE, INSERT, SELECT, UPDATE ON ` + grantEscape(schemaName) + `.* TO ` + publicUser + `; `+ "\n")
	//fmt.Println(queries.String())

	_, err = RootDB.Exec(queries.String())
	if err!=nil {
		if err := tx.Rollback(); err!=nil {
			fmt.Println(err)
			return err
		}
		return err
	}
	err = tx.Commit()
	return err
}

func DeleteProjectUser(pid int, schemaName string) error {
	tx, err := RootDB.Begin()
	if err!=nil {
		return err
	}

	internalUser := `'`+ strconv.Itoa(pid) + `_internal'@'%'`
	publicUser := `'`+ strconv.Itoa(pid) + `_public'@'%'`

	var queries strings.Builder
	queries.WriteString(`DROP USER IF EXISTS `+ internalUser +`, `+ publicUser + `; ` + "\n")

	_, err = RootDB.Exec(queries.String())
	if err!=nil {
		if err := tx.Rollback(); err!=nil {
			fmt.Println(err)
			return err
		}
		return err
	}
	err = tx.Commit()
	if err!=nil {
		if err := tx.Rollback(); err!=nil {
			fmt.Println(err)
			return err
		}
		return err
	}

	dropConn(composeProjectPublicDBUrl(pid, schemaName))
	dropConn(composeProjectInternalDBUrl(pid, schemaName))
	return nil
}
