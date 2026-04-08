package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"backend-go/api"
	"backend-go/database"
	"backend-go/notifications"
	"backend-go/scanner"
	"backend-go/storage"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

var updateChan = make(chan bool, 1)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("[Init] Using environment defaults")
	}

	database.InitDB()
	s := storage.NewFileStorage("backend-go/data/domains.json")
	api.DomainStorage = s
	scanner.DomainStorage = s

	s.OnUpdate = func(domains []storage.Domain) {
		select {
		case updateChan <- true:
		default:
		}
	}

	var activeCron *cron.Cron
	startCron := func() {
		if activeCron != nil {
			activeCron.Stop()
		}
		activeCron = cron.New()

		var intervalStr string
		database.DB.QueryRow("SELECT value FROM settings WHERE key = 'scan_interval'").Scan(&intervalStr)
		if intervalStr == "" {
			intervalStr = "24"
		}

		unit := "h"
		if intervalStr == "5" {
			unit = "m"
		}
		cronSpec := fmt.Sprintf("@every %s%s", intervalStr, unit)
		log.Printf("[Cron] Interval: %s", cronSpec)

		activeCron.AddFunc(cronSpec, func() {
			log.Printf("[Cron] Scheduled scan started")
			scanner.ScanAllDomains()
		})

		activeCron.AddFunc("0 9 * * *", func() {
			log.Println("[Cron] Daily summary task")
			notifications.SendDailySummary()
		})

		activeCron.Start()
	}

	go func() {
		for range api.CronUpdateChan {
			log.Println("[Cron] Settings updated, refreshing...")
			startCron()
		}
	}()

	startCron()
	go scanner.ScanAllDomains()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Use(cors.New())

	// Live Sync Endpoint
	app.Get("/api/events", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("Transfer-Encoding", "chunked")

		c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
			for {
				select {
				case <-updateChan:
					fmt.Fprintf(w, "data: reload\n\n")
					w.Flush()
				case <-time.After(15 * time.Second):
					fmt.Fprintf(w, ": keep-alive\n\n")
					w.Flush()
				}
			}
		})
		return nil
	})

	api.SetupRoutes(app)

	// Smart Path Resolution: Check both root and parent directory for frontend/dist
	possiblePaths := []string{
		"../frontend/dist", // When run from backend-go
		"frontend/dist",    // When run from root
	}

	distPath := ""
	for _, p := range possiblePaths {
		abs, _ := filepath.Abs(p)
		if _, err := os.Stat(filepath.Join(abs, "index.html")); err == nil {
			distPath = abs
			break
		}
	}

	if distPath == "" {
		log.Println("[Warning] frontend/dist not found! Please run 'npm run build' first.")
	} else {
		log.Printf("[Static] Serving interface from: %s", distPath)
		app.Static("/", distPath)
		app.Get("/*", func(c *fiber.Ctx) error {
			return c.SendFile(filepath.Join(distPath, "index.html"))
		})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	log.Printf("[Server] Listening on http://localhost:%s\n", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("[Server] Fatal error: %v", err)
	}
}
