package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strconv"
	"time"

	"backend-go/database"
)

type DomainRecord struct {
	ID           int
	Hostname     string
	SSLExpiry    *time.Time
	DomainExpiry *time.Time
}

func SendNotification(message string) {
	// 1. Generic Webhook
	sendWebhook(message)

	// 2. Telegram
	sendTelegram(message)

	// 3. Discord
	sendDiscord(message)

	// 4. Slack
	sendSlack(message)

	// 5. Email
	sendEmail(message)
}

func sendWebhook(message string) {
	var webhookUrl string
	database.DB.QueryRow("SELECT value FROM settings WHERE key = 'webhook_url'").Scan(&webhookUrl)
	if webhookUrl == "" {
		return
	}

	payload := map[string]string{
		"content": "🔔 **DominANT Alert**\n" + message,
	}
	jsonVal, _ := json.Marshal(payload)

	go func() {
		resp, err := http.Post(webhookUrl, "application/json", bytes.NewBuffer(jsonVal))
		if err == nil {
			defer resp.Body.Close()
		}
	}()
}

func sendTelegram(message string) {
	var token, chatID string
	database.DB.QueryRow("SELECT value FROM settings WHERE key = 'telegram_token'").Scan(&token)
	database.DB.QueryRow("SELECT value FROM settings WHERE key = 'telegram_chat_id'").Scan(&chatID)

	if token == "" || chatID == "" {
		return
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	payload := map[string]any{
		"chat_id":    chatID,
		"text":       "🔔 DominANT Alert\n" + message,
		"parse_mode": "Markdown",
	}
	jsonVal, _ := json.Marshal(payload)

	go func() {
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonVal))
		if err == nil {
			defer resp.Body.Close()
		}
	}()
}

func sendDiscord(message string) {
	var webhookUrl string
	database.DB.QueryRow("SELECT value FROM settings WHERE key = 'discord_webhook_url'").Scan(&webhookUrl)
	if webhookUrl == "" {
		return
	}

	payload := map[string]any{
		"content": "🔔 **DominANT Alert**\n" + message,
	}
	jsonVal, _ := json.Marshal(payload)

	go func() {
		resp, err := http.Post(webhookUrl, "application/json", bytes.NewBuffer(jsonVal))
		if err == nil {
			defer resp.Body.Close()
		}
	}()
}

func CheckAndNotify(domain DomainRecord) {
	var thresholdStr string
	err := database.DB.QueryRow("SELECT value FROM settings WHERE key = 'notification_threshold'").Scan(&thresholdStr)
	threshold := 30
	if err == nil && thresholdStr != "" {
		if t, e := strconv.Atoi(thresholdStr); e == nil {
			threshold = t
		}
	}

	var alerts string
	now := time.Now()

	if domain.SSLExpiry != nil {
		days := int(domain.SSLExpiry.Sub(now).Hours() / 24)
		if days >= 0 && days <= threshold {
			alerts += fmt.Sprintf("- SSL for **%s** expires in **%d days**!\n", domain.Hostname, days)
		} else if days < 0 {
			alerts += fmt.Sprintf("- SSL for **%s** has **EXPIRED**!\n", domain.Hostname)
		}
	}

	if domain.DomainExpiry != nil {
		days := int(domain.DomainExpiry.Sub(now).Hours() / 24)
		if days >= 0 && days <= threshold {
			alerts += fmt.Sprintf("- Domain registration for **%s** expires in **%d days**!\n", domain.Hostname, days)
		} else if days < 0 {
			alerts += fmt.Sprintf("- Domain registration for **%s** has **EXPIRED**!\n", domain.Hostname)
		}
	}

	if alerts != "" {
		SendNotification(alerts)
	}
}

func sendSlack(message string) {
	var webhookUrl string
	database.DB.QueryRow("SELECT value FROM settings WHERE key = 'slack_webhook_url'").Scan(&webhookUrl)
	if webhookUrl == "" {
		return
	}

	payload := map[string]any{
		"text": "🔔 *DominANT Alert*\n" + message,
	}
	jsonVal, _ := json.Marshal(payload)

	go func() {
		resp, err := http.Post(webhookUrl, "application/json", bytes.NewBuffer(jsonVal))
		if err == nil {
			defer resp.Body.Close()
		}
	}()
}

func sendEmail(message string) {
	var host, port, user, pass, from, to string
	database.DB.QueryRow("SELECT value FROM settings WHERE key = 'smtp_host'").Scan(&host)
	database.DB.QueryRow("SELECT value FROM settings WHERE key = 'smtp_port'").Scan(&port)
	database.DB.QueryRow("SELECT value FROM settings WHERE key = 'smtp_user'").Scan(&user)
	database.DB.QueryRow("SELECT value FROM settings WHERE key = 'smtp_pass'").Scan(&pass)
	database.DB.QueryRow("SELECT value FROM settings WHERE key = 'smtp_from'").Scan(&from)
	database.DB.QueryRow("SELECT value FROM settings WHERE key = 'smtp_to'").Scan(&to)

	if host == "" || user == "" || to == "" {
		return
	}

	auth := smtp.PlainAuth("", user, pass, host)
	subject := "Subject: DominANT Expiry Alert\r\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := "<html><body><h2>🔔 DominANT Alert</h2><p>" + message + "</p></body></html>"
	msg := []byte(subject + mime + body)

	go func() {
		err := smtp.SendMail(host+":"+port, auth, from, []string{to}, msg)
		if err != nil {
			fmt.Printf("[Email] Error sending: %v\n", err)
		}
	}()
}
