package handler

import (
	"encoding/json"
	"net/http"
)

type HandlerSet interface {
	Init()
}

type jsonError struct {
	Err string `json:"err"`
}

func NewJSONError(err string, status int, w http.ResponseWriter) {
	js := jsonError{err}
	j, _ := json.Marshal(js)
	http.Error(w, string(j), status)
}

func InitAllHTTPHandlers() {

	new(StaticHandler).Init()
	new(UserAPIHandler).Init()

}
