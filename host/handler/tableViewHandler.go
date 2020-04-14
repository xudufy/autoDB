package handler

import (
	"autodb/host/dbconfig"
	"autodb/host/globalsession"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
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
}

type tableInfo struct {
	Pid int
	Pname string
	Tid int
	Tname string
	DataList string
	ColumnList string
	IndexList string
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

	_ = r.ParseForm()
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

	indexRows, err := dbInternal.Query(`show indexes from ` + tInfo.Tname + ` ;`)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}
	defer indexRows.Close()

	jsI, err := dbconfig.ParseRowsToJSON(indexRows)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}
	tInfo.IndexList = string(jsI)

	err = tTmpl.Execute(w, tInfo)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}

}

func runSQLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}

	_ = r.ParseForm()
	tidS := r.Form.Get("tid")
	script := r.Form.Get("script")
	tid, err := strconv.Atoi(tidS)
	if err!= nil || script == "" {
		NewJSONError("parameter error", 400, w)
		return
	}

	tInfo := getTableInfo(tid, w, r)
	if tInfo == nil {
		return
	}

	dbPublic, err := dbconfig.GetProjectPublicConn(tInfo.Pid, tInfo.Pname)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}

	pubTx, err := dbPublic.Begin()
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}
	defer func() {
		if err!=nil {
			NewJSONError(err.Error(), 502, w)
			_ = pubTx.Rollback()
		}
	}()

	resultRows, err := pubTx.Query(script)
	if err!=nil {
		NewJSONError(err.Error(), 400, w)
		return
	}
	defer resultRows.Close()

	js, err := dbconfig.ParseRowsToJSON(resultRows)
	if err!=nil {
		return
	}

	err = pubTx.Commit()
	if err!=nil {
		return
	}

	_ = WriteJSON(js, w)
}

func addColumnHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method!="POST" {
		http.NotFound(w,r)
		return
	}

	_ = r.ParseForm()
	tidS := r.Form.Get("tid")
	tid, err := strconv.Atoi(tidS)
	if err!= nil {
		NewJSONError("parameter error", 400, w)
		return
	}

	col := ColumnInfo{
		r.Form.Get("name"),
		r.Form.Get("type"),
		r.Form.Get("options"),
	}

	err = checkColumn(col)
	if err!=nil {
		NewJSONError(err.Error(), 400, w)
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

	_, err = dbInternal.Exec(`alter table ` + tInfo.Tname +` add `+ col.Name +` ` + col.ColType + ` ` + col.Options +` ;`)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}

	return
}

func deleteColumnHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method!="POST" {
		http.NotFound(w,r)
		return
	}

	_ = r.ParseForm()
	tidS := r.Form.Get("tid")
	tid, err := strconv.Atoi(tidS)
	if err!= nil {
		NewJSONError("parameter error", 400, w)
		return
	}

	colName := r.Form.Get("name")
	ok := dbconfig.IsIdentifier(colName)
	if !ok {
		NewJSONError("column name invalid", 400, w)
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

	_, err = dbInternal.Exec(`alter table ` + tInfo.Tname +` drop column `+ colName +` ;`)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}

	return
}

func addIndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}

	_ = r.ParseForm()
	tidS := r.Form.Get("tid")
	tid, err := strconv.Atoi(tidS)
	if err!= nil {
		NewJSONError("parameter error", 400, w)
		return
	}
	indexName := r.Form.Get("name")
	columnList := r.Form.Get("columnList")
	uniqueSign := r.Form.Get("unique")
	if indexName == "" || columnList == "" || !dbconfig.IsIdentifier(indexName) {
		NewJSONError("parameter error", 400, w)
		return
	}

	if uniqueSign!="" && uniqueSign!="true" && uniqueSign!="false" {
		NewJSONError("unique should be 'true' or 'false' or omitted.", 400, w)
		return
	}

	if uniqueSign=="true" {
		uniqueSign = " unique "
	} else if uniqueSign=="false" {
		uniqueSign = ""
	}

	cols := strings.Split(columnList, ",")
	for _, v := range cols {
		if !dbconfig.IsIdentifier(strings.TrimSpace(v)) {
			NewJSONError("columnList parameter error", 400, w)
			return
		}
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

	query := `create ` + uniqueSign + ` index ` + indexName + ` on ` + tInfo.Tname + ` (` + columnList + ` );`
	fmt.Println(query)
	_, err = dbInternal.Exec(query)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}

	return
}