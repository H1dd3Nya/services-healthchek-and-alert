package alert

// Интерфейс для отправки алертов

type AlertSender interface {
	SendAlert(siteName, message string) error
}
