package handler

import (
	"autodb/host/dbconfig"
	"crypto/sha256"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
)

type UserAPIHandler struct{}

func (*UserAPIHandler) Init() {
	http.HandleFunc("/register", RegisterHandler)
}

func passwordEncode(pw string, username string) string {
	raw := "AUTODB0468091" + pw + "#HG00fh3n" + username
	after := fmt.Sprintf("%x", sha256.Sum256([]byte(raw)))
	fmt.Println(after)
	return after
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {

		t, _ := template.ParseFiles("../view/register.html")
		t.Execute(w, "")

	} else if r.Method == "POST" {

		r.ParseForm()

		um := r.PostFormValue("username")
		if m, _ := regexp.MatchString(`^([a-zA-Z0-9_\-\.]{1,100})$`, um); !m {
			NewJSONError("username invalid", 400, w)
			return
		}
		em := r.PostFormValue("email")
		if m, _ := regexp.MatchString(`^([a-zA-Z0-9_\-\.]+)@([a-zA-Z0-9_\-\.]+)\.([a-zA-Z]{2,5})$`, em); !m || len(em) >= 100 {
			NewJSONError("email invalid", 400, w)
			return
		}
		pw := r.PostFormValue("password")
		if m, _ := regexp.MatchString(`^([a-zA-Z0-9\_\.\-\+\/\=]{6,100})$`, pw); !m {
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
