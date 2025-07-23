package monitor

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

type PingResult struct {
	Timestamp time.Time
	Duration  time.Duration
	HTTPCode  int
	Success   bool
	Error     string
}

func PingHTTP(url string) PingResult {
	start := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("[PING][HTTP][FAIL] %s: %v", url, err)
		return PingResult{
			Timestamp: time.Now(),
			Duration:  time.Since(start),
			Success:   false,
			Error:     err.Error(),
		}
	}
	defer resp.Body.Close()
	res := PingResult{
		Timestamp: time.Now(),
		Duration:  time.Since(start),
		HTTPCode:  resp.StatusCode,
		Success:   resp.StatusCode >= 200 && resp.StatusCode < 400,
	}
	if res.Success {
		log.Printf("[PING][HTTP][OK] %s: %d in %v", url, res.HTTPCode, res.Duration)
	} else {
		log.Printf("[PING][HTTP][BAD] %s: %d in %v", url, res.HTTPCode, res.Duration)
	}
	return res
}

func PingTCP(address string) PingResult {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		log.Printf("[PING][TCP][FAIL] %s: %v", address, err)
		return PingResult{
			Timestamp: time.Now(),
			Duration:  time.Since(start),
			Success:   false,
			Error:     err.Error(),
		}
	}
	conn.Close()
	res := PingResult{
		Timestamp: time.Now(),
		Duration:  time.Since(start),
		Success:   true,
	}
	log.Printf("[PING][TCP][OK] %s in %v", address, res.Duration)
	return res
}

func FormatResult(siteName string, res PingResult) string {
	return fmt.Sprintf("[%s] %s: success=%v code=%d time=%v err=%s", res.Timestamp.Format(time.RFC3339), siteName, res.Success, res.HTTPCode, res.Duration, res.Error)
}
