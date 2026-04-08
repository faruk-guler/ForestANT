package scanner

import (
	"log"
	"strings"
	"sync"
	"time"

	"backend-go/database"
	"backend-go/notifications"
	"backend-go/storage"
)

var DomainStorage *storage.FileStorage

func ScanDomain(id int, hostname string) {
	domains := DomainStorage.GetAll()
	var d storage.Domain
	found := false
	for _, domain := range domains {
		if domain.ID == id {
			d = domain
			found = true
			break
		}
	}
	if !found {
		return
	}

	result := performScan(d)
	DomainStorage.Update(result)

	notifications.CheckAndNotify(notifications.DomainRecord{
		ID:           result.ID,
		Hostname:     result.Hostname,
		SSLExpiry:    result.SSLExpiry,
		DomainExpiry: result.DomainExpiry,
	})
}

func performScan(d storage.Domain) storage.Domain {
	log.Printf("[Scanner] Started: %s\n", d.Hostname)

	sslExp := GetSSLExpiry(d.Hostname)
	if sslExp != nil {
		database.LogScan(d.Hostname, "SSL", "success", "Cert valid until "+sslExp.Format("2006-01-02"))
	} else {
		database.LogScan(d.Hostname, "SSL", "error", "Failed to retrieve SSL")
	}

	whoisData := GetDomainWhoisData(d.Hostname)
	domExp := whoisData.Expiry
	nsStr := strings.Join(whoisData.Nameservers, ", ")
	availStr := "taken"
	if whoisData.Availability {
		availStr = "available"
	}

	if domExp != nil {
		database.LogScan(d.Hostname, "WHOIS", "success", "Expiry: "+domExp.Format("2006-01-02"))
	} else {
		database.LogScan(d.Hostname, "WHOIS", "error", "Failed WHOIS lookup")
	}

	securityGrade := CalculateSecurityRating(d.Hostname).Grade

	now := time.Now()
	d.Status = "active"
	d.SSLExpiry = sslExp
	d.DomainExpiry = domExp
	d.LastScan = &now
	d.Nameservers = nsStr
	d.SecurityRating = securityGrade
	d.StatusAvailability = availStr
	d.LastWhoisRaw = whoisData.Raw

	log.Printf("[Scanner] Completed: %s\n", d.Hostname)
	return d
}

func ScanAllDomains() {
	domains := DomainStorage.GetAll()
	if len(domains) == 0 {
		return
	}

	log.Printf("[Turbo] Parallel scan for %d domains (Limit: 50)\n", len(domains))

	var wg sync.WaitGroup
	sem := make(chan struct{}, 50)
	
	results := make(map[int]storage.Domain)
	var mu sync.Mutex

	for _, d := range domains {
		wg.Add(1)
		go func(domain storage.Domain) {
			defer wg.Done()
			sem <- struct{}{}
			
			res := performScan(domain)
			
			mu.Lock()
			results[res.ID] = res
			mu.Unlock()
			
			// IMMEDIATE FEEDBACK: Update memory and trigger UI reload
			DomainStorage.UpdateMemory(res)
			
			<-sem
			
			// Individual notification sent immediately
			notifications.CheckAndNotify(notifications.DomainRecord{
				ID:           res.ID,
				Hostname:     res.Hostname,
				SSLExpiry:    res.SSLExpiry,
				DomainExpiry: res.DomainExpiry,
			})
		}(d)
	}

	wg.Wait()
	
	// Batch update storage once
	DomainStorage.BatchUpdate(results)
	
	log.Println("[Turbo] All scans finished and storage updated.")
}
