package api

import (
	"net/http"
	"strings"
)

func ServeUI(uiLocation string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if strings.HasPrefix(req.URL.Path, "/api") {
			next.ServeHTTP(rw, req)
			return
		}
		http.FileServer(http.Dir(uiLocation)).ServeHTTP(rw, req)
	})
}
