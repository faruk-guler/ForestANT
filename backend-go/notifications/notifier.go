package notifications

import (
	"bytes"
	"crypto/tls"
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

// notificationSettings tüm ayarları tek sorguda tutar
// FIX: Eskiden her kanal fonksiyonu ayrı DB sorgusu yapıyordu (N+1 problemi)
type notificationSettings struct {
	WebhookURL     string
	TelegramToken  string
	TelegramChatID string
	DiscordURL     string
	SlackURL       string
	SMTPHost       string
	SMTPPort       string
	SMTPUser       string
	SMTPPass       string
	SMTPFrom       string
	SMTPTo         string
	Threshold      int
}

// loadNotificationSettings tüm ayarları TEK sorguda yükler
func loadNotificationSettings() notificationSettings {
	rows, err := database.DB.Query("SELECT key, value FROM settings")
	if err != nil {
		return notificationSettings{Threshold: 30}
	}
	defer rows.Close()

	m := make(map[string]string)
	for rows.Next() {
		var k, v string
		rows.Scan(&k, &v)
		m[k] = v
	}

	threshold := 30
	if t, err := strconv.Atoi(m["notification_threshold"]); err == nil && t > 0 {
		threshold = t
	}

	return notificationSettings{
		WebhookURL:     m["webhook_url"],
		TelegramToken:  m["telegram_token"],
		TelegramChatID: m["telegram_chat_id"],
		DiscordURL:     m["discord_webhook_url"],
		SlackURL:       m["slack_webhook_url"],
		SMTPHost:       m["smtp_host"],
		SMTPPort:       m["smtp_port"],
		SMTPUser:       m["smtp_user"],
		SMTPPass:       m["smtp_pass"],
		SMTPFrom:       m["smtp_from"],
		SMTPTo:         m["smtp_to"],
		Threshold:      threshold,
	}
}

func SendNotification(message string) {
	cfg := loadNotificationSettings()
	go sendWebhook(cfg.WebhookURL, message)
	go sendTelegram(cfg.TelegramToken, cfg.TelegramChatID, message)
	go sendDiscord(cfg.DiscordURL, message)
	go sendSlack(cfg.SlackURL, message)
	go sendEmail(cfg, message)
}

func sendWebhook(webhookURL, message string) {
	if webhookURL == "" {
		return
	}
	payload := map[string]string{
		"content": "🔔 **DominANT Alert**\n" + message,
	}
	jsonVal, _ := json.Marshal(payload)
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonVal))
	if err == nil {
		defer resp.Body.Close()
	}
}

func sendTelegram(token, chatID, message string) {
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
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonVal))
	if err == nil {
		defer resp.Body.Close()
	}
}

func sendDiscord(webhookURL, message string) {
	if webhookURL == "" {
		return
	}
	payload := map[string]any{
		"content": "🔔 **DominANT Alert**\n" + message,
	}
	jsonVal, _ := json.Marshal(payload)
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonVal))
	if err == nil {
		defer resp.Body.Close()
	}
}

func sendSlack(webhookURL, message string) {
	if webhookURL == "" {
		return
	}
	payload := map[string]any{
		"text": "🔔 *ForestANT Alert*\n" + message,
	}
	jsonVal, _ := json.Marshal(payload)
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonVal))
	if err == nil {
		defer resp.Body.Close()
	}
}

func sendEmail(cfg notificationSettings, message string) {
	if cfg.SMTPHost == "" || cfg.SMTPUser == "" || cfg.SMTPTo == "" {
		return
	}

	port := cfg.SMTPPort
	if port == "" {
		port = "587"
	}

	// FIX: TLS/STARTTLS desteği eklendi
	addr := cfg.SMTPHost + ":" + port
	auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPHost)

	subject := "Subject: ForestANT Expiry Alert\r\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := "<html><body><h2>🔔 ForestANT Alert</h2><p>" + message + "</p></body></html>"
	msg := []byte(subject + mime + body)

	from := cfg.SMTPFrom
	if from == "" {
		from = cfg.SMTPUser
	}

	// Port 465 (SMTPS) için implicit TLS
	if port == "465" {
		tlsCfg := &tls.Config{ServerName: cfg.SMTPHost}
		conn, err := tls.Dial("tcp", addr, tlsCfg)
		if err != nil {
			fmt.Printf("[Email] TLS dial error: %v\n", err)
			return
		}
		defer conn.Close()
		client, err := smtp.NewClient(conn, cfg.SMTPHost)
		if err != nil {
			fmt.Printf("[Email] SMTP client error: %v\n", err)
			return
		}
		client.Auth(auth)
		client.Mail(from)
		client.Rcpt(cfg.SMTPTo)
		w, _ := client.Data()
		w.Write(msg)
		w.Close()
		client.Quit()
		return
	}

	// Port 587 (STARTTLS)
	err := smtp.SendMail(addr, auth, from, []string{cfg.SMTPTo}, msg)
	if err != nil {
		fmt.Printf("[Email] Error sending: %v\n", err)
	}
}

func CheckAndNotify(domain DomainRecord) {
	cfg := loadNotificationSettings()

	var alerts string
	now := time.Now()

	if domain.SSLExpiry != nil {
		days := int(domain.SSLExpiry.Sub(now).Hours() / 24)
		if days >= 0 && days <= cfg.Threshold {
			alerts += fmt.Sprintf("- SSL for **%s** expires in **%d days**!\n", domain.Hostname, days)
		} else if days < 0 {
			alerts += fmt.Sprintf("- SSL for **%s** has **EXPIRED**!\n", domain.Hostname)
		}
	}

	if domain.DomainExpiry != nil {
		days := int(domain.DomainExpiry.Sub(now).Hours() / 24)
		if days >= 0 && days <= cfg.Threshold {
			alerts += fmt.Sprintf("- Domain registration for **%s** expires in **%d days**!\n", domain.Hostname, days)
		} else if days < 0 {
			alerts += fmt.Sprintf("- Domain registration for **%s** has **EXPIRED**!\n", domain.Hostname)
		}
	}

	if alerts != "" {
		SendNotification(alerts)
	}
}
