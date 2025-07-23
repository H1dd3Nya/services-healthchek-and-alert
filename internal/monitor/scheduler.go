package monitor

import (
	"log"
	"services-healthchek-and-alert/internal/alert"
	"services-healthchek-and-alert/internal/storage"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	CheckSuccess = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "service_check_success_total",
			Help: "Total successful checks",
		},
		[]string{"site_id", "type"},
	)
	CheckFail = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "service_check_fail_total",
			Help: "Total failed checks",
		},
		[]string{"site_id", "type"},
	)
	CheckDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "service_check_duration_ms",
			Help:    "Check duration in ms",
			Buckets: prometheus.LinearBuckets(10, 50, 10),
		},
		[]string{"site_id", "type"},
	)
	AlertSent = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "service_alert_sent_total",
			Help: "Total alerts sent",
		},
		[]string{"site_id"},
	)
	SitesTracked = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "service_sites_tracked",
			Help: "Number of sites being monitored",
		},
	)
)

func init() {
	prometheus.MustRegister(CheckSuccess, CheckFail, CheckDuration, AlertSent, SitesTracked)
}

type Scheduler struct {
	Sites                []storage.Site
	Results              chan SiteResult
	Stop                 chan struct{}
	Retries              map[string]int // siteID -> count
	AlertThreshold       int
	AlertSender          alert.AlertSender
	mu                   sync.Mutex
	MaxAlertsPerIncident int
	alertsSent           map[string]int           // siteID -> count
	siteMonitors         map[string]chan struct{} // siteID -> stop chan
}

type SiteResult struct {
	Site   storage.Site
	Result PingResult
}

func NewScheduler(sites []storage.Site, alertSender alert.AlertSender, alertThreshold int, maxAlertsPerIncident int) *Scheduler {
	return &Scheduler{
		Sites:                sites,
		Results:              make(chan SiteResult),
		Stop:                 make(chan struct{}),
		Retries:              make(map[string]int),
		AlertThreshold:       alertThreshold,
		AlertSender:          alertSender,
		MaxAlertsPerIncident: maxAlertsPerIncident,
		alertsSent:           make(map[string]int),
		siteMonitors:         make(map[string]chan struct{}),
	}
}

func (s *Scheduler) AddSite(site storage.Site) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, existing := range s.Sites {
		if existing.ID == site.ID {
			// Остановить старую горутину
			if stopCh, ok := s.siteMonitors[site.ID]; ok {
				close(stopCh)
			}
			// Обновить данные сайта
			s.Sites[i] = site
			// Запустить новую горутину
			ch := make(chan struct{})
			s.siteMonitors[site.ID] = ch
			go s.monitorSite(site, ch)
			return
		}
	}
	s.Sites = append(s.Sites, site)
	SitesTracked.Set(float64(len(s.Sites)))
	ch := make(chan struct{})
	s.siteMonitors[site.ID] = ch
	go s.monitorSite(site, ch)
}

func (s *Scheduler) monitorSite(site storage.Site, stopCh chan struct{}) {
	log.Printf("[INFO] Start monitoring: %s (%s)", site.Name, site.URL)
	ticker := time.NewTicker(time.Duration(site.IntervalSeconds) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			var res PingResult
			if site.CheckType == storage.CheckHTTP {
				res = PingHTTP(site.URL)
			} else {
				res = PingTCP(site.URL)
			}
			CheckDuration.WithLabelValues(site.ID, string(site.CheckType)).Observe(float64(res.Duration.Milliseconds()))
			if res.Success {
				CheckSuccess.WithLabelValues(site.ID, string(site.CheckType)).Inc()
			} else {
				CheckFail.WithLabelValues(site.ID, string(site.CheckType)).Inc()
			}
			log.Printf("[MONITOR] %s: success=%v code=%d err=%s", site.Name, res.Success, res.HTTPCode, res.Error)
			s.Results <- SiteResult{Site: site, Result: res}
			if !res.Success {
				s.Retries[site.ID]++
				log.Printf("[RETRY] %s: retries=%d (threshold=%d)", site.Name, s.Retries[site.ID], s.AlertThreshold)
				if s.Retries[site.ID] >= s.AlertThreshold && s.AlertSender != nil {
					log.Printf("[ALERT] Will call SendAlert for %s, sender=%T, retries=%d, threshold=%d", site.Name, s.AlertSender, s.Retries[site.ID], s.AlertThreshold)
					if s.alertsSent[site.ID] < s.MaxAlertsPerIncident {
						msg := FormatResult(site.Name, res)
						log.Printf("[ALERT] Sending alert for %s: %s", site.Name, msg)
						err := s.AlertSender.SendAlert(site.Name, msg)
						if err != nil {
							log.Printf("[ERROR] Alert send failed: %v", err)
						} else {
							AlertSent.WithLabelValues(site.ID).Inc()
							s.alertsSent[site.ID]++
							log.Printf("[ALERT] Alert sent for %s (%d/%d)", site.Name, s.alertsSent[site.ID], s.MaxAlertsPerIncident)
						}
					}
				}
			} else {
				s.Retries[site.ID] = 0
				s.alertsSent[site.ID] = 0
			}
		case <-stopCh:
			log.Printf("[INFO] Stop monitoring (dynamic): %s", site.Name)
			return
		case <-s.Stop:
			log.Printf("[INFO] Stop monitoring: %s", site.Name)
			return
		}
	}
}

func (s *Scheduler) Start() {
	SitesTracked.Set(float64(len(s.Sites)))
	for _, site := range s.Sites {
		ch := make(chan struct{})
		s.siteMonitors[site.ID] = ch
		go s.monitorSite(site, ch)
	}
}

func (s *Scheduler) StopAll() {
	SitesTracked.Set(0)
	close(s.Stop)
}
