package config

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
}

type TelegramConfig struct {
	BotToken string `yaml:"bot_token"`
	ChatID   string `yaml:"chat_id"`
}

type Config struct {
	SMTP                 SMTPConfig     `yaml:"smtp"`
	Telegram             TelegramConfig `yaml:"telegram"`
	AlertRetries         int            `yaml:"alert_retries"`
	MaxAlertsPerIncident int            `yaml:"max_alerts_per_incident"`
}

func LoadConfig() Config {
	cfg := Config{}
	if _, err := os.Stat("config.yaml"); err == nil {
		data, err := ioutil.ReadFile("config.yaml")
		if err != nil {
			log.Printf("[WARN] Failed to read config.yaml: %v", err)
		} else {
			err = yaml.Unmarshal(data, &cfg)
			if err != nil {
				log.Printf("[WARN] Failed to parse config.yaml: %v", err)
			}
		}
	}
	// Fallback на env, если что-то не заполнено
	if cfg.SMTP.Host == "" {
		cfg.SMTP.Host = os.Getenv("SMTP_HOST")
	}
	if cfg.SMTP.Port == "" {
		cfg.SMTP.Port = os.Getenv("SMTP_PORT")
	}
	if cfg.SMTP.Username == "" {
		cfg.SMTP.Username = os.Getenv("SMTP_USER")
	}
	if cfg.SMTP.Password == "" {
		cfg.SMTP.Password = os.Getenv("SMTP_PASS")
	}
	if cfg.SMTP.From == "" {
		cfg.SMTP.From = os.Getenv("SMTP_FROM")
	}
	if cfg.Telegram.BotToken == "" {
		cfg.Telegram.BotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	}
	if cfg.Telegram.ChatID == "" {
		cfg.Telegram.ChatID = os.Getenv("TELEGRAM_CHAT_ID")
	}
	if cfg.AlertRetries == 0 {
		if v := os.Getenv("ALERT_RETRIES"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				cfg.AlertRetries = n
			}
		}
		if cfg.AlertRetries == 0 {
			cfg.AlertRetries = 3
		}
	}
	if cfg.MaxAlertsPerIncident == 0 {
		if v := os.Getenv("MAX_ALERTS_PER_INCIDENT"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				cfg.MaxAlertsPerIncident = n
			}
		}
		if cfg.MaxAlertsPerIncident == 0 {
			cfg.MaxAlertsPerIncident = 1
		}
	}
	return cfg
}
