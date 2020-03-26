package handler

import (
	"fmt"
	"net/http"
	"os"
)

type StaticHandler struct{}

func (*StaticHandler) Init() {
	fmt.Println(os.Getwd())
	fs := http.FileServer(http.Dir("../static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
}
