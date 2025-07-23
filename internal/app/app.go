package app

import (
	"log"
	"net/http"
	"os"

	"services-healthchek-and-alert/internal/alert"
	"services-healthchek-and-alert/internal/api"
	"services-healthchek-and-alert/internal/config"
	"services-healthchek-and-alert/internal/monitor"
	"services-healthchek-and-alert/internal/storage"
	"services-healthchek-and-alert/internal/web"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Run() {
	var store storage.Storage
	if dsn := os.Getenv("DB_DSN"); dsn != "" {
		ps, err := storage.NewPostgresStorage(dsn)
		if err != nil {
			log.Fatalf("Failed to open Postgres: %v", err)
		}
		store = ps
	} else {
		bs, err := storage.NewBoltStorage("sites.db")
		if err != nil {
			log.Fatalf("Failed to open storage: %v", err)
		}
		store = bs
	}
	api.StorageInstance = store

	cfg := config.LoadConfig()
	log.Printf("[ALERT] Config.Telegram: %+v", cfg.Telegram)
	log.Printf("[ALERT] Config.SMTP: %+v", cfg.SMTP)
	alertThreshold := cfg.AlertRetries
	var alertSender alert.AlertSender
	if cfg.Telegram.BotToken != "" && cfg.Telegram.ChatID != "" {
		alertSender = &alert.TelegramAlertSender{Config: cfg.Telegram}
	} else if cfg.SMTP.Host != "" {
		alertSender = &alert.EmailAlertSender{Config: cfg.SMTP}
	}
	if alertSender == nil {
		log.Println("[ALERT] AlertSender is nil!")
	} else {
		log.Printf("[ALERT] AlertSender type: %T", alertSender)
	}

	sites, _ := store.ListSites()
	monitor.GlobalScheduler = monitor.NewScheduler(sites, alertSender, alertThreshold, cfg.MaxAlertsPerIncident)
	go monitor.GlobalScheduler.Start()
	go func() {
		for res := range monitor.GlobalScheduler.Results {
			h := storage.SiteCheckHistory{
				SiteID:    res.Site.ID,
				CheckedAt: res.Result.Timestamp.Unix(),
				Duration:  res.Result.Duration.Milliseconds(),
				HTTPCode:  res.Result.HTTPCode,
				Success:   res.Result.Success,
				Error:     res.Result.Error,
			}
			store.AddCheckHistory(h)
		}
	}()

	go func() {
		log.Println("Starting web UI on :8081...")
		http.ListenAndServe(":8081", web.NewWebServer("./web"))
	}()

	go func() {
		log.Println("Starting Prometheus metrics on :8082...")
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8082", nil)
	}()

	handler := api.NewRouter()
	log.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
