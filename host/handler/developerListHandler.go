package handler

import (
	"autodb/host/dbconfig"
	"autodb/host/globalsession"
	"database/sql"
	"html/template"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type DeveloperListHandler struct{}

var devTmpl *template.Template

func (*DeveloperListHandler) Init() {
	devTmpl, _ = template.ParseFiles("../view/developers.html")
	http.HandleFunc("/developers", viewDevelopersHandler)
	http.HandleFunc("/addDeveloper", addDeveloperHandler)
	http.HandleFunc("/deleteDeveloper", deleteDeveloperHandler)
	http.HandleFunc("/setDeveloperGroup", setDeveloperGroupHandler)
	http.HandleFunc("/searchUser", searchUserHandler)
}

//TODO: test
func viewDevelopersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method!="GET" {
		http.NotFound(w, r)
		return
	}

	err := r.ParseForm()
	if err!=nil {
		NewJSONError("parameter error", 400, w)
		return
	}

	pidS:= r.Form.Get("pid")
	if pidS== "" {
		NewJSONError( "parameter error", 400, w)
		return
	}

	pid, err := strconv.Atoi(pidS)
	if err!=nil {
		NewJSONError("parameter error", 400, w)
		return
	}

	uid, err := globalsession.GetUid(w, r)
	if err!=nil {
		NewJSONError(err.Error(), 403, w)
		return
	}

	group := globalsession.GetGroupToProject(uid, pid)
	if (group & (globalsession.UserGroupOwner | globalsession.UserGroupDeveloper))==0 {
		NewJSONError("Not Authorized", 403, w)
		return
	}

	pname:=""

	pnameRows, err := dbconfig.HostDB.Query(`select pname from projects where pid=?;`, pid)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}
	defer pnameRows.Close()

	if pnameRows.Next() {
		_ = pnameRows.Scan(&pname)
	}

	developerRows, err := dbconfig.HostDB.Query(`select U.uid, username, email, privilege 
			from (select uid, privilege from project_developer where pid=?) PD inner join users U on U.uid = PD.uid`, pid)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}
	defer developerRows.Close()

	js, err := dbconfig.ParseRowsToJSON(developerRows)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}

	tmplArgs := make(map[string]interface{})
	tmplArgs["Pid"] = pid
	tmplArgs["Pname"] = pname
	tmplArgs["List"] = string(js)
	err = devTmpl.Execute(w, tmplArgs)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}

}

func addDeveloperHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method!="POST" {
		http.NotFound(w, r)
		return
	}

	err := r.ParseForm()
	if err!=nil {
		NewJSONError("parameter error", 400, w)
		return
	}

	pidS := r.Form.Get("pid")
	pid, err := strconv.Atoi(pidS)
	if err!=nil {
		NewJSONError("parameter error", 400, w)
		return
	}

	newUidS := r.Form.Get("uid")
	newUid, err := strconv.Atoi(newUidS)
	privilege := r.Form.Get("privilege")
	if err!=nil || privilege=="" {
		NewJSONError("parameter error", 400, w)
		return
	}

	if privilege!="owner" && privilege!="developer" {
		NewJSONError("Unsupported privilege", 400, w)
		return
	}

	uid, err := globalsession.GetUid(w, r)
	if err!=nil {
		NewJSONError(err.Error(), 403, w)
		return
	}

	group := globalsession.GetGroupToProject(uid, pid)
	if (group & (globalsession.UserGroupOwner | globalsession.UserGroupDeveloper))==0 {
		NewJSONError("Not Authorized", 403, w)
		return
	}

	if group == globalsession.UserGroupDeveloper && privilege=="owner" {
		NewJSONError("Developer cannot add new owner", 403, w)
		return
	}

	_, err = dbconfig.HostDB.Exec(`insert into project_developer (uid, pid, privilege) VALUES (?,?,?)`, newUid, pid, privilege)
	if err != nil {
		NewJSONError("insert error, maybe the user has been added, or uid is invalid", 502, w)
		return
	}

}

//TODO: test
//we search exact match of uid and username (case insensitive),
//and prefix match of username if len(entry)>3.
func searchUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method!="GET" {
		http.NotFound(w, r)
		return
	}

	err := r.ParseForm()
	if err!=nil {
		NewJSONError("Parameter error", 400, w)
		return
	}

	entry := r.Form.Get("w")
	if entry=="" {
		NewJSONError("Parameter error", 400, w)
		return
	}

	uid, err := strconv.Atoi(entry)
	if err!=nil {
		uid = -1
	}

	var usersRow *sql.Rows

	if len(entry)<3 {

		usersRow, err = dbconfig.HostDB.Query(`select uid, username from users where uid = ? or UPPER(username) = UPPER(?);`, uid, entry)

	} else if m, _ :=regexp.MatchString(`^([a-zA-Z0-9_\-.]{1,100})$`, entry); m {

		var escapeEntry strings.Builder
		for _, ch := range entry {
			if ch=='_' {
				escapeEntry.WriteRune('\\')
			}
			escapeEntry.WriteRune(ch)
		}
		escapeEntry.WriteRune('%')

		usersRow, err = dbconfig.HostDB.Query(`select uid, username from users where uid = ? or UPPER(username) = UPPER(?)
			union 
			select uid, username from users where UPPER(username) like UPPER(?) LIMIT 10;`, uid, entry, escapeEntry.String())

	} else { // len(entry)>3 and entry is not a valid username
		usersRow, err = dbconfig.HostDB.Query(`select uid, username from users where uid = ?`, uid)
	}

	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}
	defer usersRow.Close()

	js, err := dbconfig.ParseRowsToJSON(usersRow)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}

	err = WriteJSON(js, w)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}

}

func deleteDeveloperHandler(w http.ResponseWriter, r *http.Request) {

}

func setDeveloperGroupHandler(w http.ResponseWriter, r *http.Request) {

}
