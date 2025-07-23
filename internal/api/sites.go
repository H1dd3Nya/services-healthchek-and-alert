package api

import (
	"encoding/json"
	"log"
	"net/http"
	"services-healthchek-and-alert/internal/storage"
	"strconv"
	"strings"

	"services-healthchek-and-alert/internal/monitor"

	"github.com/google/uuid"
)

var StorageInstance storage.Storage // Экспортируемая переменная

func handleSites(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] %s %s", r.Method, r.URL.Path)
	if StorageInstance == nil {
		log.Println("[ERROR] Storage not initialized")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("storage not initialized"))
		return
	}
	if r.Method == http.MethodPost {
		var site storage.Site
		if err := json.NewDecoder(r.Body).Decode(&site); err != nil {
			log.Printf("[ERROR] Invalid body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid body"))
			return
		}
		if site.ID == "" {
			site.ID = uuid.New().String()
			log.Printf("[INFO] Auto-generated site ID: %s", site.ID)
		}
		if err := StorageInstance.AddSite(site); err != nil {
			log.Printf("[ERROR] AddSite: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if monitor.GlobalScheduler != nil {
			monitor.GlobalScheduler.AddSite(site)
		}
		log.Printf("[INFO] Site added: %+v", site)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"id": site.ID})
		return
	}
	sites, err := StorageInstance.ListSites()
	if err != nil {
		log.Printf("[ERROR] ListSites: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	log.Printf("[INFO] ListSites: %d sites returned", len(sites))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sites)
}

func handleSiteByID(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] %s %s", r.Method, r.URL.Path)
	path := r.URL.Path[len("/sites/"):]
	if strings.HasSuffix(path, "/history") {
		id := strings.TrimSuffix(path, "/history")
		if id == "" {
			log.Println("[ERROR] History: id required")
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
			log.Printf("[ERROR] ListCheckHistory: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		log.Printf("[INFO] History returned for site %s: %d records", id, len(history))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(history)
		return
	}
	id := path
	if id == "" {
		log.Println("[ERROR] CRUD: id required")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("id required"))
		return
	}
	switch r.Method {
	case http.MethodGet:
		site, err := StorageInstance.GetSite(id)
		if err != nil {
			log.Printf("[ERROR] GetSite(%s): %v", id, err)
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(err.Error()))
			return
		}
		log.Printf("[INFO] Site returned: %+v", site)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(site)
	case http.MethodPut:
		var site storage.Site
		if err := json.NewDecoder(r.Body).Decode(&site); err != nil {
			log.Printf("[ERROR] Invalid body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid body"))
			return
		}
		if site.ID != id {
			log.Printf("[ERROR] UpdateSite: id mismatch (%s != %s)", site.ID, id)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("id mismatch"))
			return
		}
		if err := StorageInstance.AddSite(site); err != nil {
			log.Printf("[ERROR] UpdateSite: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		log.Printf("[INFO] Site updated: %+v", site)
		w.WriteHeader(http.StatusOK)
	case http.MethodDelete:
		if err := StorageInstance.DeleteSite(id); err != nil {
			log.Printf("[ERROR] DeleteSite(%s): %v", id, err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		log.Printf("[INFO] Site deleted: %s", id)
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
