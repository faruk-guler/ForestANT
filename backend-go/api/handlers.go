package api

import (
	"regexp"
	"strconv"
	"strings"
	"backend-go/acme"
	"backend-go/database"
	"backend-go/notifications"
	"backend-go/scanner"
	"backend-go/storage"
	"backend-go/workflow"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

var DomainStorage *storage.FileStorage
var CronUpdateChan = make(chan bool, 1)

type Domain = storage.Domain

var domainRegex = regexp.MustCompile(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`)

type Access struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Provider  string `json:"provider"`
	Config    string `json:"config"` // JSON string
	CreatedAt string `json:"created_at"`
}

type Workflow struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	DomainID      int     `json:"domain_id"`
	CertificateID *int    `json:"certificate_id"`
	AccessID      int     `json:"access_id"`
	Type          string  `json:"type"`
	Status        string  `json:"status"`
	LastRun       *string `json:"last_run"`
	Config        *string `json:"config"`
}

func SetupRoutes(app *fiber.App) {
	// ACME Challenge Route
	app.Get("/.well-known/acme-challenge/:token", func(c *fiber.Ctx) error {
		token := c.Params("token")
		if keyAuth, ok := acme.ActiveChallenges[token]; ok {
			return c.SendString(keyAuth)
		}
		return c.Status(404).SendString("Challenge not found")
	})

	api := app.Group("/api")

	api.Get("/dashboard", func(c *fiber.Ctx) error {
		domains := DomainStorage.GetAll()
		
		// 1. Calculate Stats in Go (Faster than React)
		healthy := 0
		critical := 0
		expired := 0
		
		now := time.Now()
		for _, d := range domains {
			sslDays := 100
			domDays := 100
			
			if d.SSLExpiry != nil {
				sslDays = int(d.SSLExpiry.Sub(now).Hours() / 24)
			}
			if d.DomainExpiry != nil {
				domDays = int(d.DomainExpiry.Sub(now).Hours() / 24)
			}

			if sslDays < 0 || domDays < 0 {
				expired++
			} else if sslDays < 30 || domDays < 30 {
				critical++
			} else {
				healthy++
			}
		}

		// 2. Get Settings
		rows, _ := database.DB.Query("SELECT key, value FROM settings")
		defer rows.Close()
		settings := make(map[string]interface{})
		for rows.Next() {
			var k, v string
			rows.Scan(&k, &v)
			settings[k] = v
		}

		return c.JSON(fiber.Map{
			"domains": domains,
			"stats": fiber.Map{
				"total":    len(domains),
				"healthy":  healthy,
				"critical": critical,
				"expired":  expired,
			},
			"settings": settings,
		})
	})

	api.Get("/domains", func(c *fiber.Ctx) error {
		domains := DomainStorage.GetAll()
		return c.JSON(domains)
	})

	api.Post("/domains", func(c *fiber.Ctx) error {
		payload := struct {
			Hostname string `json:"hostname"`
		}{}
		if err := c.BodyParser(&payload); err != nil || payload.Hostname == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Hostname is required"})
		}

		hostname := strings.ToLower(strings.TrimSpace(payload.Hostname))
		if !domainRegex.MatchString(hostname) {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid hostname format"})
		}

		// Duplicate check
		for _, d := range DomainStorage.GetAll() {
			if d.Hostname == hostname {
				return c.Status(409).JSON(fiber.Map{"error": "Domain already exists"})
			}
		}

		id := DomainStorage.Add(storage.Domain{
			Hostname: hostname,
			Status:   "pending",
		})

		go scanner.ScanDomain(id, hostname)
		return c.JSON(fiber.Map{"id": id, "hostname": hostname, "status": "pending"})
	})

	api.Post("/domains/bulk", func(c *fiber.Ctx) error {
		payload := struct {
			Domains []string `json:"domains"`
		}{}
		
		if err := c.BodyParser(&payload); err != nil {
			msg := fmt.Sprintf("Bulk import failed: Invalid JSON body or format: %v", err)
			database.LogScan("SYSTEM", "IMPORT", "error", msg)
			return c.Status(400).JSON(fiber.Map{"error": msg})
		}
		
		if len(payload.Domains) == 0 {
			msg := "Bulk import failed: No domains provided in the request"
			database.LogScan("SYSTEM", "IMPORT", "warning", msg)
			return c.Status(400).JSON(fiber.Map{"error": msg})
		}

		allDomains := DomainStorage.GetAll()
		var toAdd []storage.Domain
		skippedCount := 0
		duplicateCount := 0
		
		for _, hostname := range payload.Domains {
			hostname = strings.ToLower(strings.TrimSpace(hostname))
			if hostname == "" || !domainRegex.MatchString(hostname) {
				if hostname != "" {
					database.LogScan(hostname, "IMPORT", "warning", "Skipped: Invalid hostname format")
				}
				skippedCount++
				continue
			}

			exists := false
			for _, d := range allDomains {
				if d.Hostname == hostname {
					exists = true
					break
				}
			}

			if exists {
				duplicateCount++
				continue
			}

			// Prevent duplicates within the same bulk upload
			duplicateInPayload := false
			for _, ta := range toAdd {
				if ta.Hostname == hostname {
					duplicateInPayload = true
					break
				}
			}
			
			if !duplicateInPayload {
				toAdd = append(toAdd, storage.Domain{Hostname: hostname, Status: "pending"})
			} else {
				duplicateCount++
			}
		}

		added := 0
		if len(toAdd) > 0 {
			added = DomainStorage.AddBulk(toAdd)
			go scanner.ScanAllDomains()
		}

		summary := fmt.Sprintf("Bulk import completed. Added: %d, Duplicates: %d, Invalid: %d", added, duplicateCount, skippedCount)
		database.LogScan("SYSTEM", "IMPORT", "info", summary)
		
		return c.JSON(fiber.Map{
			"message": summary,
			"added": added,
			"duplicates": duplicateCount,
			"invalid": skippedCount,
			"total_scanned": len(payload.Domains),
		})
	})

	api.Post("/domains/bulk-delete", func(c *fiber.Ctx) error {
		payload := struct {
			Ids []int `json:"ids"`
		}{}
		if err := c.BodyParser(&payload); err != nil || len(payload.Ids) == 0 {
			return c.Status(400).JSON(fiber.Map{"error": "Array of ids is required"})
		}

		for _, id := range payload.Ids {
			DomainStorage.Delete(id)
		}
		return c.JSON(fiber.Map{"message": "Bulk delete completed"})
	})

	api.Put("/domains/:id", func(c *fiber.Ctx) error {
		id, _ := strconv.Atoi(c.Params("id"))
		payload := struct {
			Hostname string `json:"hostname"`
		}{}
		if err := c.BodyParser(&payload); err != nil || payload.Hostname == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Hostname is required"})
		}

		hostname := strings.ToLower(strings.TrimSpace(payload.Hostname))
		
		all := DomainStorage.GetAll()
		var target storage.Domain
		found := false
		for _, v := range all {
			if v.ID == id {
				target = v
				found = true
				break
			}
		}

		if !found {
			return c.Status(404).JSON(fiber.Map{"error": "Domain not found"})
		}

		target.Hostname = hostname
		target.Status = "pending"
		target.SSLExpiry = nil
		target.DomainExpiry = nil
		target.LastScan = nil

		DomainStorage.Update(target)
		go scanner.ScanDomain(id, hostname)
		return c.JSON(fiber.Map{"message": "Domain updated and re-scan queued"})
	})

	api.Delete("/domains/:id", func(c *fiber.Ctx) error {
		id, _ := strconv.Atoi(c.Params("id"))
		DomainStorage.Delete(id)
		return c.SendStatus(204)
	})

	api.Post("/scan", func(c *fiber.Ctx) error {
		go scanner.ScanAllDomains()
		return c.JSON(fiber.Map{"message": "Scan started in the background."})
	})

	api.Post("/scan/:id", func(c *fiber.Ctx) error {
		id, _ := strconv.Atoi(c.Params("id"))
		all := DomainStorage.GetAll()
		for _, v := range all {
			if v.ID == id {
				go scanner.ScanDomain(id, v.Hostname)
				return c.JSON(fiber.Map{"message": "Domain scan started in the background."})
			}
		}
		return c.Status(404).SendString("Not found")
	})

	api.Get("/settings", func(c *fiber.Ctx) error {
		rows, err := database.DB.Query("SELECT key, value FROM settings")
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		defer rows.Close()

		settings := make(map[string]interface{})
		for rows.Next() {
			var k, v string
			rows.Scan(&k, &v)
			settings[k] = v
		}
		return c.JSON(settings)
	})

	api.Post("/settings", func(c *fiber.Ctx) error {
		payload := make(map[string]string)
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid body"})
		}

		for k, v := range payload {
			database.DB.Exec("UPDATE settings SET value = ? WHERE key = ?", v, k)
		}

		// Signal cron refresh
		select {
		case CronUpdateChan <- true:
		default:
		}

		return c.JSON(fiber.Map{"message": "Settings updated"})
	})

	api.Post("/settings/test-webhook", func(c *fiber.Ctx) error {
		notifications.SendNotification("✅ **Test Başarılı!** DominANT Go Webhook bağlantınız düzgün çalışıyor.")
		return c.JSON(fiber.Map{"message": "Test message sent. Please check your webhook channel."})
	})

	api.Get("/logs", func(c *fiber.Ctx) error {
		rows, err := database.DB.Query("SELECT id, hostname, type, status, message, created_at FROM scan_logs ORDER BY created_at DESC LIMIT 100")
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		defer rows.Close()

		var logs []map[string]interface{}
		for rows.Next() {
			var id int
			var hostname, scanType, status, message, createdAt string
			rows.Scan(&id, &hostname, &scanType, &status, &message, &createdAt)
			logs = append(logs, map[string]interface{}{
				"id":         id,
				"hostname":   hostname,
				"type":       scanType,
				"status":     status,
				"message":    message,
				"created_at": createdAt,
			})
		}
		if logs == nil {
			logs = []map[string]interface{}{}
		}
		return c.JSON(logs)
	})

	api.Delete("/logs", func(c *fiber.Ctx) error {
		_, err := database.DB.Exec("DELETE FROM scan_logs")
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to clear logs"})
		}
		return c.SendStatus(204)
	})

	// ACCESS ROUTES
	api.Get("/access", func(c *fiber.Ctx) error {
		rows, err := database.DB.Query("SELECT id, name, provider, config, created_at FROM access")
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		defer rows.Close()

		var list []Access
		for rows.Next() {
			var a Access
			if err := rows.Scan(&a.ID, &a.Name, &a.Provider, &a.Config, &a.CreatedAt); err == nil {
				list = append(list, a)
			}
		}
		if list == nil {
			list = []Access{}
		}
		return c.JSON(list)
	})

	api.Post("/access", func(c *fiber.Ctx) error {
		var a Access
		if err := c.BodyParser(&a); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}
		_, err := database.DB.Exec("INSERT INTO access (name, provider, config) VALUES (?, ?, ?)", a.Name, a.Provider, a.Config)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(201).JSON(fiber.Map{"message": "Access created"})
	})

	api.Delete("/access/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		_, err := database.DB.Exec("DELETE FROM access WHERE id = ?", id)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"message": "Access deleted"})
	})

	// WORKFLOW ROUTES
	api.Get("/workflows", func(c *fiber.Ctx) error {
		rows, err := database.DB.Query("SELECT id, name, domain_id, certificate_id, access_id, type, status, last_run, config FROM workflows")
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		defer rows.Close()

		var list []Workflow
		for rows.Next() {
			var w Workflow
			if err := rows.Scan(&w.ID, &w.Name, &w.DomainID, &w.CertificateID, &w.AccessID, &w.Type, &w.Status, &w.LastRun, &w.Config); err == nil {
				list = append(list, w)
			}
		}
		if list == nil {
			list = []Workflow{}
		}
		return c.JSON(list)
	})

	api.Post("/workflows", func(c *fiber.Ctx) error {
		var w Workflow
		if err := c.BodyParser(&w); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}
		_, err := database.DB.Exec("INSERT INTO workflows (name, domain_id, access_id, type, config) VALUES (?, ?, ?, ?, ?)", w.Name, w.DomainID, w.AccessID, w.Type, w.Config)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(201).JSON(fiber.Map{"message": "Workflow created"})
	})

	api.Post("/workflows/:id/run", func(c *fiber.Ctx) error {
		id, _ := c.ParamsInt("id")
		go workflow.RunWorkflow(id)
		return c.JSON(fiber.Map{"message": "Workflow execution started in the background."})
	})
}
