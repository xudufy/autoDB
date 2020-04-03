package handler

import (
	"autodb/host/dbconfig"
	"autodb/host/globalsession"
	"crypto/sha256"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
)

type UserAPIHandler struct{}

var (
	loginTmpl *template.Template
	logoutTmpl *template.Template
	registerTmpl *template.Template
)

func (*UserAPIHandler) Init() {
	loginTmpl, _ = template.ParseFiles("../view/login.html")
	logoutTmpl, _ = template.ParseFiles("../view/logout.html")
	registerTmpl, _ = template.ParseFiles("../view/register.html")
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
}

func passwordEncode(pw string, username string) string {
	raw := "AUTODB0468091" + pw + "#HG00fh3n" + username
	after := fmt.Sprintf("%x", sha256.Sum256([]byte(raw)))
	fmt.Println(after)
	return after
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		_ = registerTmpl.Execute(w, "")
	} else if r.Method == "POST" {

		_ = r.ParseForm()

		um := r.PostFormValue("username")
		if m, _ := regexp.MatchString(`^([a-zA-Z0-9_\-.]{1,100})$`, um); !m {
			NewJSONError("username invalid", 400, w)
			return
		}
		em := r.PostFormValue("email")
		if m, _ := regexp.MatchString(`^([a-zA-Z0-9_\-.]+)@([a-zA-Z0-9_\-.]+)\.([a-zA-Z]{2,5})$`, em); !m || len(em) >= 100 {
			NewJSONError("email invalid", 400, w)
			return
		}
		pw := r.PostFormValue("password")
		if m, _ := regexp.MatchString(`^([a-zA-Z0-9_.\-+/=]{6,100})$`, pw); !m {
			NewJSONError("password invalid", 400, w)
			return
		}

		arg := map[string]interface{}{
			"um": um,
			"em": em,
			"pw": passwordEncode(pw, um),
		}

		query, args, err := dbconfig.PrepareNamedQuery("INSERT INTO users (username,email,pw) values (:um, :em, :pw)", arg)
		if err != nil {
			NewJSONError("register sql:"+err.Error(), 400, w)
			return
		}

		result, err := dbconfig.HostDB.Exec(query, args...)
		if err != nil {
			NewJSONError("register sql:"+err.Error(), 400, w)
			return
		}
		fmt.Println(result.RowsAffected())

	} else {
		http.NotFound(w, r)
	}

}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method=="GET" {
		_ = loginTmpl.Execute(w, "")
	} else if r.Method=="POST" {
		_ = r.ParseForm()
		em:= r.FormValue("email")
		pw:=r.FormValue("password")
		if em=="" || pw== ""{
			NewJSONError("Arguments cannot be empty.", 400, w)
			return
		}

		rows, err := dbconfig.HostDB.Query("select uid, username, pw from users where email=?", em)
		if err!=nil {
			NewJSONError("login sql:"+err.Error(), 400, w)
			return
		}
		defer rows.Close()
		var realPw, realUm string
		var realUid int
		rows.Next()
		err = rows.Scan(&realUid, &realUm, &realPw)
		if err!=nil {
			fmt.Println(err.Error())
			http.Error(w, "internal login error", 502)
			return
		}
		if rows.Next() {
			fmt.Println("login query return multiple results.")
			http.Error(w, "internal login error", 502)
			return
		}

		if realPw==passwordEncode(pw, realUm) {
			sess, _ := globalsession.GSess.SessionStart(w, r)
			defer sess.SessionRelease(w)
			_ = sess.Set("uid", realUid)
			return
		} else {
			globalsession.GSess.SessionDestroy(w, r)
		}

	} else {
		http.NotFound(w, r)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		globalsession.GSess.SessionDestroy(w, r)
		_ = logoutTmpl.Execute(w, "")
	} else {
		http.NotFound(w, r)
	}
}
