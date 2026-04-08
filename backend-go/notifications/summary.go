package notifications

import (
	"fmt"
	"log"
	"time"

	"backend-go/database"
)

func SendDailySummary() {
	log.Println("[Summary] Generating daily summary report...")

	rows, err := database.DB.Query("SELECT hostname, ssl_expiry, domain_expiry, status FROM domains")
	if err != nil {
		log.Println("[Summary] Error querying domains:", err)
		return
	}
	defer rows.Close()

	var healthy, critical, expired int
	var report string
	now := time.Now()

	for rows.Next() {
		var hostname, status string
		var sslExp, domExp *string
		if err := rows.Scan(&hostname, &sslExp, &domExp, &status); err != nil {
			continue
		}

		isCritical := false
		isExpired := false

		if sslExp != nil {
			t, _ := time.Parse("2006-01-02T15:04:05Z", *sslExp)
			days := int(t.Sub(now).Hours() / 24)
			if days < 0 {
				isExpired = true
			} else if days < 30 {
				isCritical = true
			}
		}

		if domExp != nil {
			t, _ := time.Parse("2006-01-02T15:04:05Z", *domExp)
			days := int(t.Sub(now).Hours() / 24)
			if days < 0 {
				isExpired = true
			} else if days < 30 {
				isCritical = true
			}
		}

		if isExpired {
			expired++
		} else if isCritical {
			critical++
		} else {
			healthy++
		}
	}

	total := healthy + critical + expired
	if total == 0 {
		return
	}

	report = fmt.Sprintf("📊 **Daily Summary Report**\n\n- Totals: **%d** domains\n- ✅ Healthy: **%d**\n- ⚠️ Critical: **%d**\n- ❌ Expired: **%d**\n\nKeep monitoring with DominANT!", total, healthy, critical, expired)
	
	SendNotification(report)
}
