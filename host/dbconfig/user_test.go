package dbconfig

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestAddProjectUser(t *testing.T) {
	Init()
	err := AddProjectUser(1, "developer_project_example1")
	if err!=nil {
		t.Errorf(err.Error())
	}

	dbinternal,err := GetProjectInternalConn(1, "developer_project_example1")
	if err!=nil {
		t.Errorf(err.Error())
		return
	}

	row,err := dbinternal.Query("show tables from autodb")
	if err==nil {
		t.Errorf("project_internal should not have rights to visit autodb")
		fmt.Println(ParseRowsToJSON(row))
	} else {
		fmt.Println("PASS:" + err.Error())
	}

	row,err = dbinternal.Query("show tables")
	if err!=nil {
		t.Errorf(err.Error())
		return
	} else {
		fmt.Println(ParseRowsToJSON(row))
		row.Close()
	}

	dbpublic, err := GetProjectPublicConn(1, "developer_project_example1")
	if err!=nil {
		t.Errorf(err.Error())
		return
	}
	row, err = dbpublic.Query("show tables from autodb;")
	if err==nil {
		t.Errorf("project_public should not have rights to visit autodb")
		fmt.Println(ParseRowsToJSON(row))
	} else {
		fmt.Println("PASS:" + err.Error())
	}

	row,err = dbpublic.Query("show tables;")
	if err!=nil {
		t.Errorf(err.Error())
	} else {
		fmt.Println(ParseRowsToJSON(row))
		row.Close()
	}

	_, err = dbpublic.Exec("create table aaa(aaab int primary key)")
	if err==nil {
		t.Errorf("project_public should not have rights to create table")
	} else {
		fmt.Println("PASS:" + err.Error())
	}

}

func TestJsonUn(t *testing.T) {
	inputForm := make(map[string]interface{})
	bodyInBytes := []byte(`{"a":1, "b": "b"}`)
	err := json.Unmarshal(bodyInBytes, &inputForm)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(inputForm)
}

func TestQueryInsert(t *testing.T) {
	Init()
	rows, err := HostDB.Query("insert into nullTest (id, time) values (2, '2020-03-29 08:00:00');")
	if err!=nil {
		t.Errorf(err.Error())
		return
	}
	js, _ := ParseRowsToJSON(rows)
	fmt.Println(js)
}