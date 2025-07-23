package alert

import (
	"fmt"
	"log"
	"net/smtp"
	"services-healthchek-and-alert/internal/config"
)

type EmailAlertSender struct {
	Config config.SMTPConfig
}

func (e *EmailAlertSender) SendAlert(siteName, message string) error {
	auth := smtp.PlainAuth("", e.Config.Username, e.Config.Password, e.Config.Host)
	addr := fmt.Sprintf("%s:%s", e.Config.Host, e.Config.Port)
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: ALERT: %s\r\n\r\n%s", e.Config.From, siteName, message))
	err := smtp.SendMail(addr, auth, e.Config.From, []string{e.Config.From}, msg)
	if err != nil {
		log.Printf("[ERROR] Email alert send failed: %v", err)
	} else {
		log.Printf("[ALERT] Email alert sent: %s", siteName)
	}
	return err
}
