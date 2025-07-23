package api

import (
	"net/http"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/sites", handleSites)
	mux.HandleFunc("/sites/", handleSiteByID)
	return mux
}
