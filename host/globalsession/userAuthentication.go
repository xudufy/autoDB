package globalsession

import (
	"encoding/json"

	"github.com/astaxie/beego/session"
)

var GlobalSessions *session.Manager

func Init() {
	var cfg session.ManagerConfig
	json.Unmarshal([]byte(`{"cookieName":"__autodb", "enableSetCookie,omitempty": true, "gclifetime":3600, "maxLifetime": 3600, "secure": false, "cookieLifeTime": 3600, "providerConfig": ""}`), &cfg)
	GlobalSessions, _ = session.NewManager("memory", &cfg)
	go GlobalSessions.GC()
}
