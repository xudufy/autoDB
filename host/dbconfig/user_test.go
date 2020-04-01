package dbconfig

import (
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