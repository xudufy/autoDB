package handler

import (
	"autodb/host/dbconfig"
	"autodb/host/globalsession"
	"html/template"
	"net/http"
	"strconv"
)

type TableViewHandler struct{}

var tTmpl *template.Template

func (*TableViewHandler) Init() {
	tTmpl, _ = template.ParseFiles("../view/table.html")
	http.HandleFunc("/table", viewTableHandler)
	http.HandleFunc("/runScript", runSQLHandler)
	http.HandleFunc("/addColumn", addColumnHandler)
	http.HandleFunc("/deleteColumn", deleteColumnHandler)
	http.HandleFunc("/addIndex", addIndexHandler)
	http.HandleFunc("/deleteIndex", deleteIndexHandler)
}

type tableInfo struct {
	Pid int
	Pname string
	Tid int
	Tname string
	DataList string
	ColumnList string
}

//find pid pname tid tname use a given tid, and do user authentication.
//return nil indicate error, and error is handled inside.
func getTableInfo(tid int, w http.ResponseWriter, r *http.Request) *tableInfo {
	var err error
	r1, err := dbconfig.HostDB.Query(`select pid, pname, tid, name from (select pid, tid, name from tables where tid = ?) A inner join projects P on A.pid = P.pid`, tid)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return nil
	}
	defer r1.Close()

	var tInfo tableInfo
	if r1.Next() {
		err = r1.Scan(&tInfo.Pid, &tInfo.Pname, &tInfo.Tid, &tInfo.Tname)
		if err!=nil {
			NewJSONError(err.Error(), 502, w)
			return nil
		}
	} else {
		http.NotFound(w, r)
		return nil
	}

	uid, err := globalsession.GetUid(w, r)
	if err!=nil {
		NewJSONError(err.Error(), 403, w)
		return nil
	}

	group:=globalsession.GetGroupToProject(uid, tInfo.Pid)
	if (group & (globalsession.UserGroupDeveloper | globalsession.UserGroupOwner)) == 0 {
		NewJSONError("Not Authorized", 403, w)
		return nil
	}

	return &tInfo
}

func viewTableHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method!="GET" {
		http.NotFound(w, r)
		return
	}

	_ := r.ParseForm()
	tidS := r.Form.Get("tid")
	tid, err := strconv.Atoi(tidS)
	if err!= nil {
		NewJSONError("parameter error", 400, w)
		return
	}

	tInfo := getTableInfo(tid, w, r)
	if tInfo == nil {
		return
	}

	dbInternal, err := dbconfig.GetProjectInternalConn(tInfo.Pid, tInfo.Pname)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}

	ColDataRows, err := dbInternal.Query(`describe `+tInfo.Tname+`;`)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}
	defer ColDataRows.Close()

	js, err := dbconfig.ParseRowsToJSON(ColDataRows)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}
	tInfo.ColumnList = string(js)

	sampleDataRows, err := dbInternal.Query(`select * from ` + tInfo.Tname + ` LIMIT 20;`)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}
	defer sampleDataRows.Close()

	jsD, err := dbconfig.ParseRowsToJSON(sampleDataRows)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}
	tInfo.DataList = string(jsD)

	err = tTmpl.Execute(w, tInfo)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}

}

func runSQLHandler(w http.ResponseWriter, r *http.Request) {

}

func addColumnHandler(w http.ResponseWriter, r *http.Request) {

}

func deleteColumnHandler(w http.ResponseWriter, r *http.Request) {

}

func addIndexHandler(w http.ResponseWriter, r *http.Request) {

}

func deleteIndexHandler(w http.ResponseWriter, r *http.Request) {

}