package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
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

func WriteJSON(js []byte, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(js)))

	jsReader := bytes.NewReader(js)
	_, err := io.Copy(w, jsReader)
	return err
}

func InitAllHTTPHandlers() {

	new(StaticHandler).Init()
	new(UserAPIHandler).Init()
	new(GenericAPIHandler).Init()
	new(ProjectListHandler).Init()
	new(TableListHandler).Init()
	new(DeveloperListHandler).Init()
	new(TableViewHandler).Init()
	new(ApiViewHandler).Init()

}

func CROSWrapper(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Cookies, Content-Type, Content-Length, Accept-Encoding")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "-1") // use -1 because we response different origin according to the origin in request.
			if r.Method=="OPTIONS" {
				return
			}
			h.ServeHTTP(w, r)
			return
		})
}
