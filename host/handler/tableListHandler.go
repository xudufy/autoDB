package handler

import (
	"autodb/host/dbconfig"
	"autodb/host/globalsession"
	"html/template"
	"net/http"
	"strconv"
)

type TableListHandler struct {}

var tableListTmpl *template.Template

func (*TableListHandler) Init() {
	tableListTmpl, _ = template.ParseFiles("../view/tables.html")
	http.HandleFunc("/project", viewTableListHandler)
	http.HandleFunc("/addTable", addTableHandler)
}

func viewTableListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method!="GET" {
		http.NotFound(w, r)
		return
	}

	err := r.ParseForm()
	if err!=nil {
		NewJSONError("parameter error", 400, w)
		return
	}
	pidS := r.FormValue("pid")
	if pidS=="" {
		http.NotFound(w, r)
		return
	}

	pid, err := strconv.Atoi(pidS)
	if err!=nil {
		http.NotFound(w, r)
		return
	}

	uid, err := globalsession.GetUid(w, r)
	if err!=nil {
		NewJSONError(err.Error(), 403, w)
		return
	}

	group := globalsession.GetGroupToProject(uid, pid)
	if (group & (globalsession.UserGroupDeveloper | globalsession.UserGroupOwner)) == 0 {
		NewJSONError("Not your project.", 403, w)
		return
	}

	tableRows, err := dbconfig.HostDB.Query(`select tid, name from tables where pid = ?;`, pid)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}
	defer tableRows.Close()

	js, err := dbconfig.ParseRowsToJSON(tableRows)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}

	err = tableListTmpl.Execute(w, string(js))
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}

}

func addTableHandler(w http.ResponseWriter, r *http.Request) {

}