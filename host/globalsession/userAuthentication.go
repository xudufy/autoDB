package globalsession

import (
	"autodb/host/dbconfig"
	"errors"
	"net/http"

	"autodb/host/session"
)

var GSess *session.Manager
var CookieName = "__autodb"


func Init() {
	cfg := new(session.ManagerConfig)
	*cfg = session.ManagerConfig{
		CookieName:CookieName,
		EnableSetCookie: true,
		DisableHTTPOnly: false,
		Gclifetime:3600,
		Maxlifetime: 3600,
		Secure: false,
		CookieLifeTime: 3600,
		ProviderConfig: "",
		SameSite: http.SameSiteNoneMode,
	}
	GSess, _ = session.NewManager("memory", cfg)
	go GSess.GC()
}

const (
	UserGroupOwner = 1 << iota
	UserGroupDeveloper
	UserGroupUser
	UserGroupOther
)

func UserGroupToString(group int) string {
	switch group {
	case UserGroupUser:
		return "user"
	case UserGroupDeveloper:
		return "developer"
	case UserGroupOwner:
		return "developer"
	case UserGroupOther:
		return "other"
	}
	return ""
}

//will return -1, err if not logged in.
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

//uid<0 indicate not logged in.
func GetGroupToProject (uid int, pid int) int {

	if uid<0 {
		return UserGroupOther
	}

	devRow, err := dbconfig.HostDB.Query("select privilege from project_developer where uid = ? and pid = ?;", uid, pid)
	if err!= nil {
		panic(err) //dbconnection issue.
		return UserGroupUser
	}
	defer devRow.Close()

	var privilege string
	if devRow.Next() {
		_ = devRow.Scan(&privilege)
	}

	if privilege=="owner" {
		return UserGroupOwner
	} else if privilege=="developer" {
		return UserGroupDeveloper
	}

	return UserGroupUser
}
