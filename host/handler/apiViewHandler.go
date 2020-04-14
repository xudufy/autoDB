package handler

import (
	"html/template"
	"net/http"
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

}

type projectInfo struct {
	pid int
	pname string
}

func addApiHandler(w http.ResponseWriter, r *http.Request) {

}

func modifyApiHandler(w http.ResponseWriter, r *http.Request) {

}

func deleteApiHandler(w http.ResponseWriter, r *http.Request) {

}