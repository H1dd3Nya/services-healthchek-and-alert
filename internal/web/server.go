package web

import (
	"net/http"
)

func NewWebServer(staticDir string) http.Handler {
	return http.FileServer(http.Dir(staticDir))
}
