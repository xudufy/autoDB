package handler

import (
	"autodb/host/dbconfig"
	"autodb/host/globalsession"
	"encoding/json"
	"html/template"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type TableListHandler struct {}

var tableListTmpl *template.Template

func (*TableListHandler) Init() {
	tableListTmpl, _ = template.ParseFiles("../view/tables.html")
	http.HandleFunc("/project", viewTableListHandler)
	http.HandleFunc("/addTable/", addTableHandler)
	http.HandleFunc("/deleteTable", deleteTableHandler)
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

type ColumnInfo struct {
	Name string `json:"name"`
	ColType string `json:"type"`
	Options string `json:"options"`
}

type TableInfo struct {
	Name string `json:"name"`
	Columns []ColumnInfo `json:"columns"`
	Options string `json:"options"`
}


func addTableHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method!="POST" {
		http.NotFound(w, r)
		return
	}

	urlPart := strings.Split(r.URL.Path, "/")[1:]
	if len(urlPart)!=2 || urlPart[1]=="" {
		http.NotFound(w, r)
		return
	}

	pid, err := strconv.Atoi(urlPart[1])
	if err!=nil {
		http.NotFound(w, r)
		return
	}

	if r.ContentLength > 8192 {
		NewJSONError("too long", 400, w)
		return
	}

	uid, err := globalsession.GetUid(w, r)
	if err!=nil {
		NewJSONError("Not Authorized", 403, w)
		return
	}
	group:= globalsession.GetGroupToProject(uid, pid)
	if (group & (globalsession.UserGroupOwner | globalsession.UserGroupDeveloper)) == 0 {
		NewJSONError("Not Authorized",403, w)
		return
	}

	pname := ""
	pnameRow, err := dbconfig.HostDB.Query(`select pname from projects where pid=?;`, pid)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}
	defer pnameRow.Close()
	if pnameRow.Next() {
		_ = pnameRow.Scan(&pname)
	}

	if pname == "" {
		NewJSONError("project not found", 502, w)
		return
	}

	var form TableInfo

	err = json.NewDecoder(r.Body).Decode(&form)
	if err!=nil {
		NewJSONError("json input parse error", 400, w)
		return
	}

	//fmt.Println(form)

	if ok, _ := regexp.MatchString(`^[A-Za-z_][\w_]{0,63}$`, form.Name); !ok {
		NewJSONError("table name invalid", 400, w)
		return
	}

	for _, v := range form.Columns {
		if ok, _ := regexp.MatchString(`^[A-Za-z_][\w_]{0,63}$`, v.Name); !ok {
			NewJSONError("column name" + v.Name + " invalid", 400, w)
			return
		}

		if ok, _ := regexp.MatchString(`^[^,;]*$`, v.Options); !ok {
			NewJSONError("column options should not contain , and ;", 400, w)
			return
		}

		if ok := checkColumnType(v.ColType);!ok {
			NewJSONError( "column type unsupported", 400, w)
			return
		}
	}

	if ok, _ := regexp.MatchString(`^[^'";]*$`, form.Options); !ok {
		NewJSONError(`table option must not contain ' " and ;`, 400, w)
		return
	}

	var filteredOptions strings.Builder
	filteredOptions.Reset()
	parenthesesLevel := 0
	for _, ch := range form.Options {
		switch ch {
		case '(':
			parenthesesLevel++
		case ')':
			parenthesesLevel--
			if parenthesesLevel<0 {
				NewJSONError(`table option parentheses do not paired`, 400, w)
				return
			}
		case ',':
			if parenthesesLevel>0 {
				filteredOptions.WriteRune(';') // we already tested that there is no ';' in form.Options
				continue
			}
		}
		filteredOptions.WriteRune(ch)
	}

	options := strings.Split(filteredOptions.String(), ",")
	for _, vo := range options {
		v := strings.TrimSpace(vo)
		flag := (v=="") || strings.HasPrefix(v, "PRIMARY KEY")
		flag = flag || strings.HasPrefix(v, "INDEX")
		flag = flag || strings.HasPrefix(v, "UNIQUE")
		flag = flag || strings.HasPrefix(v, "FOREIGN KEY")
		if !flag {
			NewJSONError("table option unsupported: "+ vo, 400, w)
			return
		}
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

	_, err = tx.Exec(`insert into tables (pid, name) VALUES (?, ?)`, pid, form.Name)
	if err!=nil {
		NewJSONError("name has been used", 400, w)
		return
	}

	var tableBuildQuery strings.Builder
	tableBuildQuery.WriteString(`create table ` + form.Name + " (\n")

	for i, v:=range form.Columns {
		if i!=0 {
			tableBuildQuery.WriteRune(',')
		}
		tableBuildQuery.WriteString(v.Name+" "+ v.ColType + " " + v.Options + "\n")
	}

	form.Options = strings.TrimSpace(form.Options)
	if form.Options != "" {
		tableBuildQuery.WriteString("," + form.Options)
	}

	tableBuildQuery.WriteString(");")

	dbc, err := dbconfig.GetProjectInternalConn(pid, pname)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}
	_, err = dbc.Exec(tableBuildQuery.String())
	if err!=nil {
		NewJSONError(err.Error(), 400, w)
		return
	}

	err = tx.Commit()

}

func checkColumnType(coltype string) bool {
	coltype = strings.TrimSpace(coltype)
	switch coltype {
	case "MEDIUMTEXT":
		fallthrough
	case "INT":
		fallthrough
	case "BIGINT":
		fallthrough
	case "DOUBLE":
		fallthrough
	case "DATETIME":
		return true
	}

	if ok, _ := regexp.MatchString(`^VARCHAR\([0-9]{1,5}\)$`, coltype); ok {
		return true
	}

	if ok, _ := regexp.MatchString(`^ENUM\(\s*'[\w_ ]+'\s*(,\s*'[\w_ ]+'\s*)+\)$`, coltype); ok {
		return true
	}

	return false
}

func deleteTableHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method!="POST" {
		http.NotFound(w, r)
		return
	}

	err := r.ParseForm()
	if err!=nil {
		NewJSONError("parameter error", 400, w)
		return
	}
	tidS := r.FormValue("tid")
	if tidS=="" {
		http.NotFound(w, r)
		return
	}
	tid, err := strconv.Atoi(tidS)
	if err!=nil {
		http.NotFound(w, r)
		return
	}

	pRow, err := dbconfig.HostDB.Query(`select P.pid, pname, T.name from (select pid, name from tables where tid = ?) T inner join projects P on P.pid = T.pid;`, tid)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}
	defer pRow.Close()

	pid := -1
	pname := ""
	tname := ""
	if pRow.Next() {
		_ = pRow.Scan(&pid, &pname, &tname)
	}
	if pid==-1 || pname == "" || tname == ""{
		NewJSONError(`Table not found`, 404, w)
		return
	}

	uid, err := globalsession.GetUid(w, r)
	if err!=nil {
		NewJSONError(err.Error(), 403, w)
		return
	}
	group:= globalsession.GetGroupToProject(uid, pid)
	if (group & (globalsession.UserGroupOwner | globalsession.UserGroupDeveloper)) == 0 {
		NewJSONError("Not Authorized",403, w)
		return
	}

	inConn, err := dbconfig.GetProjectInternalConn(pid, pname)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}

	_, err = inConn.Exec(`drop table ` + tname + `;`)
	if err!=nil {
		NewJSONError(`cannot drop the table ` + err.Error(), 502, w)
		return
	}

	_, err = dbconfig.HostDB.Exec(`delete from tables where tid = ?`, tid)
	if err!=nil {
		NewJSONError(`Table has been deleted, but metadata update failed. A metadata sync may fix the problem.`, 502, w)
		return
	}

}