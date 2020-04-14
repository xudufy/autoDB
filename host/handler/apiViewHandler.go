package handler

import (
	"autodb/host/dbconfig"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

type ApiViewHandler struct{}

var apiTmpl *template.Template

func (*ApiViewHandler) Init() {
	apiTmpl, _ = template.ParseFiles("../view/apis.html")
	http.HandleFunc("/apis", viewApiHandler)
	http.HandleFunc("/addApi", addApiHandler)
	http.HandleFunc("/modifyApi", modifyApiHandler)
	http.HandleFunc("/deleteApi", deleteApiHandler)
}

func viewApiHandler(w http.ResponseWriter, r *http.Request) {
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

	apisRows, err := dbconfig.HostDB.Query(`select name, aid, type, tmpl from apis where tid = ?`, tid)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}
	defer apisRows.Close()

	js, err := dbconfig.ParseRowsToJSON(apisRows)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}

	viewData := map[string]interface{} {
		"Tid":tid,
		"Tname": tInfo.Tname,
		"ApiList": string(js),
	}

	err = apiTmpl.Execute(w, viewData)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}

}

type apiInfo struct {
	aid string
	name string
	tmpl string
	scope string
}

func addApiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method!="POST" {
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
	
	aInfo := apiInfo{
		name:  r.Form.Get("name"),
		tmpl:  r.Form.Get("script"),
		scope: r.Form.Get("type"),
	}

	if !dbconfig.IsIdentifier(aInfo.name) || aInfo.tmpl=="" {
		NewJSONError("parameter error", 400, w)
		return
	}

	switch aInfo.scope {
	case "public":
		aInfo.scope = "public"
	case "user":
		aInfo.scope = "user-domain"
	case "developer":
		aInfo.scope = "developer-domain"
	default:
		NewJSONError("api type error", 400, w)
		return
	}

	_, _, err = dbconfig.ParseNamedQuery(aInfo.tmpl)
	if err!=nil {
		NewJSONError(err.Error(), 400, w)
		return
	}

	tInfo := getTableInfo(tid, w, r)
	if tInfo==nil {
		return
	}

	entropy := make([]byte, 64)
	if _, err = rand.Read(entropy); err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}

	aInfo.aid = fmt.Sprintf("%x", sha256.Sum256(entropy))

	_, err = dbconfig.HostDB.Exec(`insert into apis (aid, tid, name, type, tmpl) values (?,?,?,?,?)`,
		aInfo.aid, tid, aInfo.name, aInfo.scope, aInfo.tmpl)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}

}

func modifyApiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method!="POST" {
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

	aInfo := apiInfo{
		name:  strings.TrimSpace(r.Form.Get("name")),
		tmpl:  strings.TrimSpace(r.Form.Get("script")),
		scope: strings.TrimSpace(r.Form.Get("type")),
	}

	if !dbconfig.IsIdentifier(aInfo.name) {
		NewJSONError("parameter error", 400, w)
		return
	}

	switch aInfo.scope {
	case "public":
		aInfo.scope = "public"
	case "user":
		aInfo.scope = "user-domain"
	case "developer":
		aInfo.scope = "developer-domain"
	case "":
		aInfo.scope = ""
	default:
		NewJSONError("api type error", 400, w)
		return
	}

	if aInfo.tmpl != "" {
		_, _, err = dbconfig.ParseNamedQuery(aInfo.tmpl)
		if err!=nil {
			NewJSONError(err.Error(), 400, w)
			return
		}
	}

	tInfo := getTableInfo(tid, w, r)
	if tInfo==nil {
		return
	}

	if aInfo.scope == "" && aInfo.tmpl == "" {
		NewJSONError("Nothing to modify", 400, w)
		return
	} else if aInfo.scope=="" {
		_, err := dbconfig.HostDB.Exec(`update apis set tmpl=? where tid=? and name=?`, aInfo.tmpl, tid, aInfo.name)
		if err!=nil {
			NewJSONError(err.Error(), 502, w)
			return
		}
	} else if aInfo.tmpl== "" {
		_, err := dbconfig.HostDB.Exec(`update apis set type=? where tid=? and name=?`, aInfo.scope, tid, aInfo.name)
		if err!=nil {
			NewJSONError(err.Error(), 502, w)
			return
		}
	} else {
		_, err := dbconfig.HostDB.Exec(`update apis set type=?, tmpl = ? where tid=? and name=?`, aInfo.scope, aInfo.tmpl, tid, aInfo.name)
		if err!=nil {
			NewJSONError(err.Error(), 502, w)
			return
		}
	}

}

func deleteApiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method!="POST" {
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

	name := strings.TrimSpace(r.Form.Get("name"))

	if !dbconfig.IsIdentifier(name) {
		NewJSONError("parameter error", 400, w)
		return
	}

	tInfo := getTableInfo(tid, w, r)
	if tInfo == nil {
		return
	}

	_, err = dbconfig.HostDB.Exec(`delete from apis where tid = ? and name = ?;`, tid, name)
	if err!=nil {
		NewJSONError(err.Error(), 502, w)
		return
	}
}