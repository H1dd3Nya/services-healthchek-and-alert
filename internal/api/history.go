package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func handleSiteHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		return
	}
	path := r.URL.Path
	if !strings.HasSuffix(path, "/history") {
		return
	}
	id := strings.TrimSuffix(strings.TrimPrefix(path, "/sites/"), "/history")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("id required"))
		return
	}
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil {
			limit = n
		}
	}
	history, err := StorageInstance.ListCheckHistory(id, limit)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}
