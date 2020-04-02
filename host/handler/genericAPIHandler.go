package handler

import (
	"autodb/host/dbconfig"
	"autodb/host/globalsession"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type GenericAPIHandler struct {}

func (*GenericAPIHandler) Init() {
	http.HandleFunc("/api/", apiHandler)
}

func filterTypePrefixInForm(inputForm map[string]interface{}) error {
	for k := range inputForm {
		if k[:5] == "time_" {
			_, ok := inputForm[k].(string)
			if !ok {
				return errors.New(k+" is not a RFC3339 time")
			}
			thisTime, err := time.Parse(time.RFC3339, inputForm[k].(string))
			if err!=nil {
				return errors.New(k+" is not a RFC3339 time")
			}
			inputForm[k] = thisTime.In(time.Local).Format("2006-01-02 15:04:05")
		}
	}
	return nil
}

func readJSONFormInBody(w http.ResponseWriter, r *http.Request) (map[string]interface{}, error) {
	bodyInBytes := make([]byte, r.ContentLength*2) //make sure we can read the entire body.
	ret, err := r.Body.Read(bodyInBytes)
	bodyInBytes = bodyInBytes[:ret+1]
	if err!=io.EOF {
		if err==nil {
			test := make([]byte, 1)
			ret, err := r.Body.Read(test)
			if ret!=0 || err!=io.EOF {
				NewJSONError("http body too long2", 400, w)
				return nil, errors.New("form error")
			}
		} else {
			NewJSONError(err.Error(), 502, w)
			return nil, errors.New("form error")
		}
	}
	inputForm := make(map[string]interface{})
	err = json.Unmarshal(bodyInBytes, &inputForm)
	if err != nil {
		NewJSONError("parameters are not in JSON format", 400, w)
		return nil, errors.New("form error")
	}
	return inputForm, nil
}

//TODO: test
func apiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method!="POST" {
		http.NotFound(w, r)
		return
	}
	if r.ContentLength>4096000 {
		NewJSONError("http body too long", 400, w)
		return
	}

	urlPart := strings.Split(r.URL.Path, "/")[1:]
	if len(urlPart)!=2 || urlPart[1]=="" {
		http.NotFound(w, r)
		return
	}
	aid := urlPart[1]

	apiInfoRow, err := dbconfig.HostDB.Query(`select aid, tid, type, tmpl from apis where aid = ?`, aid)
	if err!=nil {
		NewJSONError(err.Error(),502, w)
		return
	}
	defer apiInfoRow.Close()
	if !apiInfoRow.Next() {
		http.NotFound(w, r)
		return
	}

	var (
		tid int
		apiType string
		tmpl string
	)

	err = apiInfoRow.Scan(&aid, &tid, &apiType, &tmpl)
	if err!=nil {
		NewJSONError(err.Error(),502, w)
		return
	}

	pRow, err := dbconfig.HostDB.Query(`
		select A.pid, pname
		from 
		(select pid from tables where tid=?) A
		inner join projects P on A.pid=P.pid;
	`, tid)
	if err!=nil {
		NewJSONError(err.Error(),502, w)
		return
	}
	defer pRow.Close()
	if !pRow.Next() {
		NewJSONError("api: cannot find the project id",502, w)
		return
	}
	var (
		pid int
		pname string
	)
	err = pRow.Scan(&pid, &pname)
	if err!=nil {
		NewJSONError(err.Error(),502, w)
		return
	}

	dbGuest, err := dbconfig.GetProjectPublicConn(pid, pname)
	if err!= nil {
		NewJSONError(err.Error(),502, w)
		return
	}

	uid := -1
	if apiType!="public" {
		uidTemp, group := globalsession.GetUserIdAndGroupToProject(pid, w, r)
		if group==globalsession.UserGroupOther {
			NewJSONError("login required", 403, w)
			return
		}
		uid = uidTemp
		if apiType=="developer-domain" && group != globalsession.UserGroupOwner && group != globalsession.UserGroupDeveloper {
			NewJSONError("developer login required", 403, w)
			return
		}
	}

	inputForm, _ := readJSONFormInBody(w, r)
	if inputForm==nil {
		return // error message handled inside readJSONFormInBody
	}
	inputForm["currentUserID"] = uid

	err = filterTypePrefixInForm(inputForm)
	if err!=nil {
		NewJSONError(err.Error(), 400, w)
		return
	}

	query, args, err := dbconfig.PrepareNamedQuery(tmpl, inputForm)
	if err!=nil {
		NewJSONError(err.Error(), 400, w)
		return
	}

	rows, err:= dbGuest.Query(query, args...)
	if err!=nil {
		NewJSONError("api execution error", 502, w)
		fmt.Println("api execution error:" + err.Error())
		return
	}
	defer rows.Close()

	js, err := dbconfig.ParseRowsToJSON(rows)
	if err!=nil {
		NewJSONError(err.Error(), 502, w) //should not happen.
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(js)))
	
	jsReader := bytes.NewReader(js)
	_, err = io.Copy(w, jsReader)

	if err!=nil {
		NewJSONError(err.Error(), 502, w) //should not happen.
		return
	}

}