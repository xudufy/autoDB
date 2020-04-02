package globalsession

import (
	"autodb/host/dbconfig"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/astaxie/beego/session"
)

var GSess *session.Manager

func Init() {
	var cfg session.ManagerConfig
	json.Unmarshal([]byte(`{
		"cookieName":"__autodb", 
		"enableSetCookie,omitempty": true, 
		"gclifetime":3600, 
		"maxLifetime": 3600, 
		"secure": false, 
		"cookieLifeTime": 3600, 
		"providerConfig": ""}`), &cfg)
	GSess, _ = session.NewManager("memory", &cfg)
	go GSess.GC()
}

const (
	UserGroupOwner = iota
	UserGroupDeveloper
	UserGroupUser
	UserGroupOther
)

func GetUid(w http.ResponseWriter, r *http.Request) (int, error) {
	sess, _ := GSess.SessionStart(w, r)
	defer sess.SessionRelease(w)
	uidI := sess.Get("uid")
	uid, ok := uidI.(int)
	if !ok || uid<0 {
		return -1, errors.New("not logged in")
	}
	return uid, nil
}

func GetUserIdAndGroupToProject (pid int, w http.ResponseWriter, r *http.Request) (int, int) {
	uid, _ := GetUid(w,r)
	if uid<0 {
		return uid, UserGroupOther
	}

	devRow, err := dbconfig.HostDB.Query("select privilege from project_developer where uid = ? and pid = ?;", uid, pid)
	if err!= nil {
		panic(err) //dbconnection issue.
		return uid, UserGroupUser
	}
	defer devRow.Close()

	var privilege string
	if devRow.Next() {
		devRow.Scan(&privilege)
	}

	if privilege=="owner" {
		return uid, UserGroupOwner
	} else if privilege=="developer" {
		return uid, UserGroupDeveloper
	}

	return uid, UserGroupUser
}
