package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"services-healthchek-and-alert/internal/config"
)

type TelegramAlertSender struct {
	Config config.TelegramConfig
}

func (t *TelegramAlertSender) SendAlert(siteName, message string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.Config.BotToken)
	body := map[string]string{
		"chat_id": t.Config.ChatID,
		"text":    fmt.Sprintf("ALERT: %s\n%s", siteName, message),
	}
	jsonBody, _ := json.Marshal(body)
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Printf("[ERROR] Telegram alert send failed: %v", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		log.Printf("[ERROR] Telegram alert send failed: %s", resp.Status)
		return fmt.Errorf("telegram error: %s", resp.Status)
	}
	log.Printf("[ALERT] Telegram alert sent: %s", siteName)
	return nil
}
