package scanner

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"
	"net"
)

func isInternalDomain(hostname string) bool {
	// 1. Check if it's an IP address
	if net.ParseIP(hostname) != nil {
		return true
	}

	// 2. Check for common internal/non-public TLDs
	internalTLDs := []string{".local", ".lan", ".home", ".corp", ".internal", ".test", ".invalid", ".localhost", ".intranet"}
	hostname = strings.ToLower(hostname)
	for _, tld := range internalTLDs {
		if strings.HasSuffix(hostname, tld) {
			return true
		}
	}

	// 3. No dots usually means local mDNS or host shortcut
	if !strings.Contains(hostname, ".") {
		return true
	}

	return false
}

func getDomainExpiryRDAP(domain string) *time.Time {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", "https://rdap.org/domain/"+domain, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var data struct {
		Events []struct {
			EventAction string `json:"eventAction"`
			EventDate   string `json:"eventDate"`
		} `json:"events"`
	}
	if err := json.Unmarshal(body, &data); err == nil {
		for _, e := range data.Events {
			if e.EventAction == "expiration" {
				if t, parseErr := time.Parse(time.RFC3339, e.EventDate); parseErr == nil {
					return &t
				}
			}
		}
	}
	return nil
}

func getDomainExpiryNetworkCalc(domain string) *time.Time {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", "https://networkcalc.com/api/dns/whois/"+domain, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var data struct {
		Status string `json:"status"`
		Whois  struct {
			Expires                string `json:"expires"`
			RegistryExpirationDate string `json:"registry_expiration_date"`
		} `json:"whois"`
	}
	if err := json.Unmarshal(body, &data); err == nil && data.Status == "OK" {
		expStr := data.Whois.RegistryExpirationDate
		if expStr == "" {
			expStr = data.Whois.Expires
		}
		if expStr != "" {
			if t, parseErr := time.Parse(time.RFC3339, expStr); parseErr == nil {
				return &t
			}
			// Alternative simple layout
			if t, parseErr := time.Parse("2006-01-02T15:04:05Z", expStr); parseErr == nil {
				return &t
			}
		}
	}
	return nil
}

func getDomainExpiryTRFallback(domain string) *time.Time {
	if !strings.HasSuffix(domain, ".tr") {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", "https://whois.enis.org.tr/whois?domain="+domain, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// Improved regex for .tr expiration
	re := regexp.MustCompile(`(?i)(Expires on|Bitiş Tarihi)\.+:\s*([0-9]{4}-[0-9]{2}-[0-9]{2})`)
	match := re.FindStringSubmatch(bodyStr)
	if len(match) > 2 {
		if t, parseErr := time.Parse("2006-01-02", match[2]); parseErr == nil {
			return &t
		}
	}
	return nil
}

type WhoisData struct {
	Expiry       *time.Time
	Nameservers  []string
	Availability bool // True if available for registration
	Raw          string
}

func GetDomainWhoisData(hostname string) WhoisData {
	data := WhoisData{Availability: false, Raw: "On-Prem / Internal Asset"}
	
	if isInternalDomain(hostname) {
		return data
	}

	// Extract root domain more robustly
	domain := hostname
	parts := strings.Split(hostname, ".")
	if len(parts) >= 3 {
		last := parts[len(parts)-1]
		secondLast := parts[len(parts)-2]
		// Common multi-part TLDs (e.g., .com.tr, .co.uk, .org.uk)
		if len(secondLast) <= 3 && (last == "tr" || last == "uk" || last == "au" || last == "nz") {
			domain = strings.Join(parts[len(parts)-3:], ".")
		} else {
			domain = strings.Join(parts[len(parts)-2:], ".")
		}
	} else if len(parts) == 2 {
		domain = strings.Join(parts, ".")
	}

	// Native WHOIS library is best for NS and Raw data
	result, err := whois.Whois(domain)
	if err == nil {
		data.Raw = result
		parsed, err := whoisparser.Parse(result)
		if err == nil && parsed.Domain != nil {
			// Expiry
			if parsed.Domain.ExpirationDate != "" {
				if t, parseErr := time.Parse(time.RFC3339, parsed.Domain.ExpirationDate); parseErr == nil {
					data.Expiry = &t
				}
			}
			// Nameservers
			data.Nameservers = parsed.Domain.NameServers
			// Availability Check
			status := strings.Join(parsed.Domain.Status, " ")
			status = strings.ToLower(status)
			if strings.Contains(status, "not found") || strings.Contains(status, "no match") || strings.Contains(status, "available") {
				data.Availability = true
			}
		}
	}

	// Fallback for Expiry only (via RDAP/NetworkCalc for better accuracy)
	if data.Expiry == nil {
		exp := getDomainExpiryRDAP(domain)
		if exp == nil {
			exp = getDomainExpiryNetworkCalc(domain)
		}
		if exp == nil {
			exp = getDomainExpiryTRFallback(domain)
		}
		data.Expiry = exp
	}

	return data
}
