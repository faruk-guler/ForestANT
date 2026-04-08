package storage

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Domain struct {
	ID                int        `json:"id"`
	Hostname          string     `json:"hostname"`
	SSLExpiry         *time.Time `json:"ssl_expiry"`
	DomainExpiry      *time.Time `json:"domain_expiry"`
	LastScan          *time.Time `json:"last_scan"`
	Status            string     `json:"status"`
	Nameservers       string     `json:"nameservers"`
	SecurityRating    string     `json:"security_rating"`
	StatusAvailability string     `json:"status_availability"`
	LastWhoisRaw      string     `json:"last_whois_raw"`
}

type FileStorage struct {
	FilePath string
	Domains  []Domain
	Mu       sync.RWMutex
	OnUpdate func([]Domain)
}

func NewFileStorage(path string) *FileStorage {
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Printf("[Storage] Warning: Could not resolve absolute path for %s: %v", path, err)
		absPath = path // fallback to relative
	}
	fs := &FileStorage{FilePath: absPath}
	fs.Load()
	go fs.Watch()
	return fs
}

func (fs *FileStorage) Load() error {
	fs.Mu.Lock()
	defer fs.Mu.Unlock()

	data, err := os.ReadFile(fs.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			fs.Domains = []Domain{}
			return nil
		}
		return err
	}

	if len(data) == 0 {
		fs.Domains = []Domain{}
		return nil
	}

	err = json.Unmarshal(data, &fs.Domains)
	if err != nil {
		log.Printf("[Storage] Error unmarshaling: %v", err)
		return err
	}

	return nil
}

func (fs *FileStorage) Save() error {
	fs.Mu.Lock()
	defer fs.Mu.Unlock()
	return fs.saveLocked()
}

func (fs *FileStorage) saveLocked() error {
	data, err := json.MarshalIndent(fs.Domains, "", "  ")
	if err != nil {
		return err
	}

	// Ensure parent directory exists
	dir := filepath.Dir(fs.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Atomic write: write to .tmp, then rename
	tmpPath := fs.FilePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmpPath, fs.FilePath)
}

func (fs *FileStorage) Watch() {
	var lastMod time.Time
	for {
		info, err := os.Stat(fs.FilePath)
		if err == nil {
			if info.ModTime().After(lastMod) {
				if lastMod.IsZero() {
					lastMod = info.ModTime()
				} else {
					log.Println("[Storage] File change detected externally, reloading...")
					lastMod = info.ModTime()
					fs.Load()
					if fs.OnUpdate != nil {
						fs.OnUpdate(fs.Domains)
					}
				}
			}
		}
		time.Sleep(2 * time.Second)
	}
}

func (fs *FileStorage) GetAll() []Domain {
	fs.Mu.RLock()
	defer fs.Mu.RUnlock()
	
	// Return a copy to prevent external mutation of the slice
	res := make([]Domain, len(fs.Domains))
	copy(res, fs.Domains)
	return res
}

func (fs *FileStorage) Add(domain Domain) int {
	fs.Mu.Lock()
	defer fs.Mu.Unlock()

	maxID := 0
	for _, d := range fs.Domains {
		if d.ID > maxID {
			maxID = d.ID
		}
	}
	domain.ID = maxID + 1
	// Prepend to the slice so it appears at the top
	fs.Domains = append([]Domain{domain}, fs.Domains...)
	
	if err := fs.saveLocked(); err != nil {
		log.Printf("[Storage] Error saving after add: %v", err)
	}
	return domain.ID
}

func (fs *FileStorage) AddBulk(domains []Domain) int {
	if len(domains) == 0 {
		return 0
	}
	fs.Mu.Lock()
	defer fs.Mu.Unlock()

	maxID := 0
	for _, d := range fs.Domains {
		if d.ID > maxID {
			maxID = d.ID
		}
	}
	
	addedCount := 0
	for i := range domains {
		maxID++
		domains[i].ID = maxID
		fs.Domains = append(fs.Domains, domains[i])
		addedCount++
	}
	
	if err := fs.saveLocked(); err != nil {
		log.Printf("[Storage] Error saving after bulk add: %v", err)
	}
	return addedCount
}

func (fs *FileStorage) Update(domain Domain) {
	fs.Mu.Lock()
	defer fs.Mu.Unlock()

	found := false
	for i, d := range fs.Domains {
		if d.ID == domain.ID {
			fs.Domains[i] = domain
			found = true
			break
		}
	}
	
	if found {
		if err := fs.saveLocked(); err != nil {
			log.Printf("[Storage] Error saving after update: %v", err)
		}
	}
}

func (fs *FileStorage) UpdateMemory(domain Domain) {
	fs.Mu.Lock()
	defer fs.Mu.Unlock()

	found := false
	for i, d := range fs.Domains {
		if d.ID == domain.ID {
			fs.Domains[i] = domain
			found = true
			break
		}
	}

	if found && fs.OnUpdate != nil {
		fs.OnUpdate(fs.Domains)
	}
}

func (fs *FileStorage) BatchUpdate(updates map[int]Domain) {
	fs.Mu.Lock()
	defer fs.Mu.Unlock()

	updated := false
	for i, d := range fs.Domains {
		if up, ok := updates[d.ID]; ok {
			fs.Domains[i] = up
			updated = true
		}
	}

	if updated {
		if err := fs.saveLocked(); err != nil {
			log.Printf("[Storage] Error saving after batch update: %v", err)
		}
	}
}

func (fs *FileStorage) ReplaceAll(domains []Domain) {
	fs.Mu.Lock()
	defer fs.Mu.Unlock()

	fs.Domains = domains
	if err := fs.saveLocked(); err != nil {
		log.Printf("[Storage] Error saving after replace all: %v", err)
	}
}

func (fs *FileStorage) Delete(id int) {
	fs.Mu.Lock()
	defer fs.Mu.Unlock()

	newDomains := []Domain{}
	for _, d := range fs.Domains {
		if d.ID != id {
			newDomains = append(newDomains, d)
		}
	}
	fs.Domains = newDomains
	
	if err := fs.saveLocked(); err != nil {
		log.Printf("[Storage] Error saving after delete: %v", err)
	}
}
