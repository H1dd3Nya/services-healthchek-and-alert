package storage

type CheckType string

const (
	CheckHTTP CheckType = "http"
	CheckTCP  CheckType = "tcp"
)

type Site struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	URL             string    `json:"url"`
	CheckType       CheckType `json:"check_type"`
	IntervalSeconds int       `json:"interval_seconds"`
}

type SiteCheckHistory struct {
	SiteID    string `json:"site_id"`
	CheckedAt int64  `json:"checked_at"`
	Duration  int64  `json:"duration_ms"`
	HTTPCode  int    `json:"http_code,omitempty"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}
