package handler

import (
	"autodb/host/dbconfig"
	"autodb/host/globalsession"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strconv"
)

type ProjectListHandler struct{}

var projectsTmpl *template.Template

func (*ProjectListHandler) Init() {
	projectsTmpl, _ = template.ParseFiles("../view/projects.html")
	http.HandleFunc("/projects", projectListHandler)
	http.HandleFunc("/createProject", createProjectHandler)
	http.HandleFunc("/deleteProject", deleteProjectHandler)
}

func projectListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.NotFound(w, r)
		return
	}
	uid, err := globalsession.GetUid(w, r)
	if err!=nil {
		NewJSONError("Not logged in", 403, w)
		return
	}

	pRows, err := dbconfig.HostDB.Query(
		`select A.pid, pname, create_time 
				from (select pid from project_developer where uid=?) A 
                    inner join projects P on A.pid=P.pid;`, uid)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}
	defer pRows.Close()

	//We do not necessarily need to write the complex html templates.
	//We can put a json into template as a variable in javascript.
	//and use the js to render the html on client side.
	js, err := dbconfig.ParseRowsToJSON(pRows)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}

	err = projectsTmpl.Execute(w, string(js))
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}

}

//insert a info row into autodb.projects.
//insert a info row into autodb.project_developer
//create a database with projects.pname
//create a user with projects.pid+”_public” and a pw
//grant the user DELETE INSERT SELECT UPDATE privilege on the new database.
//create a user with projects.pid+”_internal” and a pw
//grant the user ALL privilege on the new database.
func createProjectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}

	var (
		uid int
		pid int
		pname string
	)

	uidT, err :=globalsession.GetUid(w, r)
	if err!= nil {
		NewJSONError("Login Required.", 403, w)
		return
	}
	uid = uidT

	err = r.ParseForm()
	if err!=nil {
		NewJSONError("Parameter Error", 400, w)
		return
	}

	pname = r.Form.Get("name")

	if ok, err:=regexp.MatchString(`^[A-Za-z_][\w_]{0,63}$`, pname); err!=nil || !ok{
		NewJSONError("Parameter Error", 400, w)
		return
	}

	tx, err := dbconfig.HostDB.Begin()
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}
	defer func() {
		if err!=nil {
			_ = tx.Rollback()
		}
	}()

	result, err := tx.Exec(`insert into projects (pname) values (?);`, pname)
	if err!=nil {
		NewJSONError("project name has been used", 502, w)
		fmt.Println(err)
		return
	}

	pidi64, err := result.LastInsertId()
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}
	pid = int(pidi64)

	_, err = tx.Exec(`insert into project_developer (uid, pid, privilege) values (?, ?, 'owner');`, uid, pid)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}

	_, err = tx.Exec(`CREATE DATABASE ` +pname+ `;`)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}

	err = tx.Commit()
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}

	err = dbconfig.AddProjectUser(pid, pname)
	if err!=nil {
		NewJSONError(err.Error(), 502, w) // this should not happen
		err = nil
		return
	}

}

// I have a feel that this will be buggy
func deleteProjectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method!="POST" {
		http.NotFound(w, r)
		return
	}

	var (
		uid int
		pid int
		pname string
	)
	var err error = nil

	uid, err = globalsession.GetUid(w, r)
	if err!=nil {
		NewJSONError(err.Error(), 403, w)
		return
	}

	err = r.ParseForm()
	if err!=nil {
		NewJSONError("parameter error", 400, w)
		return
	}

	pidS:= r.Form.Get("pid")
	if pidS=="" {
		NewJSONError("parameter error", 400, w)
		return
	}

	pid, err = strconv.Atoi(pidS)
	if err!=nil {
		NewJSONError("parameter error", 400, w)
		return
	}

	group := globalsession.GetGroupToProject(uid, pid)
	if group != globalsession.UserGroupOwner {
		NewJSONError("Only owner can delete project.", 403, w)
		return
	}

	rows, err := dbconfig.HostDB.Query(`select pname from projects where pid=?;`, pid)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}
	defer rows.Close()

	if rows.Next() {
		_ = rows.Scan(&pname)
	} else {
		NewJSONError("pid does not exist", 400, w)
		return
	}

	err = dbconfig.DeleteProjectUser(pid, pname)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}
	defer func() {
		if err!=nil {
			_ = dbconfig.AddProjectUser(pid, pname)
		}
	}()

	_, err = dbconfig.RootDB.Exec(`drop database if exists ` + pname + `;`)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}

	_, err = dbconfig.HostDB.Exec(`delete from projects where pid=?`, pid)
	if err!=nil {
		NewJSONError(`Project has been deleted, but metadata update failed. A metadata sync may fix the problem.`, 502, w)
		return
	}

}

