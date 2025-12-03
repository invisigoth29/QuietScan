package vendordb

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"quietscan/assets"
)

type OUIEntry struct {
	Prefix string `json:"prefix"`
	Vendor string `json:"vendor"`
}

type OUIDatabase map[string]string

type OUIDiff struct {
	NewVendors     []string
	UpdatedVendors []string
	RemovedVendors []string
	TotalMaster    int
	TotalLocal     int
}

func LoadLocalOUI() (OUIDatabase, error) {
	// First, try to load from disk next to executable (if user has updated the database)
	// This allows users to update the database and have it persist
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		diskPath := filepath.Join(exeDir, "assets", "oui.json")
		if data, err := os.ReadFile(diskPath); err == nil {
			var db OUIDatabase
			if err := json.Unmarshal(data, &db); err == nil && len(db) > 0 {
				return db, nil
			}
		}
	}

	// Fallback: try current directory (for development)
	if data, err := os.ReadFile("assets/oui.json"); err == nil {
		var db OUIDatabase
		if err := json.Unmarshal(data, &db); err == nil && len(db) > 0 {
			return db, nil
		}
	}

	// Finally, use bundled resource (always available)
	data := assets.ResourceOuiJson.Content()
	var db OUIDatabase
	if err := json.Unmarshal(data, &db); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bundled OUI data: %w", err)
	}

	return db, nil
}

func FetchMasterOUI() (OUIDatabase, string, error) {
	// Try multiple endpoints; IEEE sometimes moves/blocks URLs.
	candidates := []string{
		"https://standards-oui.ieee.org/oui/oui.json",
		"https://standards-oui.ieee.org/oui/oui.csv",
		"https://standards-oui.ieee.org/oui/oui.txt",
		"https://regauth.standards.ieee.org/standards-ra-web/rest/assignments/oui/public.json",
		"https://raw.githubusercontent.com/wireshark/wireshark/master/manuf",
	}

	if override := strings.TrimSpace(os.Getenv("QUIETSCAN_OUI_URL")); override != "" {
		candidates = append([]string{override}, candidates...)
	}

	client := &http.Client{Timeout: 20 * time.Second}
	var errs []string

	for _, url := range candidates {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: build request: %v", url, err))
			continue
		}
		req.Header.Set("User-Agent", "QuietScan/1.0 (+https://github.com/donovanfarrell/quietscan)")
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", url, err))
			continue
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: read body: %v", url, err))
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			trimmed := strings.TrimSpace(string(body))
			if len(trimmed) > 200 {
				trimmed = trimmed[:200] + "..."
			}
			errs = append(errs, fmt.Sprintf("%s: status %d body=%q", url, resp.StatusCode, trimmed))
			continue
		}

		if db, err := parseOUIJSON(body); err == nil {
			return db, url, nil
		}
		if db, err := parseOUICSV(body); err == nil {
			return db, url, nil
		}
		if db, err := parseOUITXT(body); err == nil {
			return db, url, nil
		}
		if db, err := parseWiresharkManuf(body); err == nil {
			return db, url, nil
		}
		errs = append(errs, fmt.Sprintf("%s: unrecognized format", url))
	}

	// If all remote attempts fail, fall back to local bundled data to avoid blocking updates entirely.
	if local, err := LoadLocalOUI(); err == nil {
		return local, "assets/oui.json", nil
	}

	return nil, "", fmt.Errorf("failed to fetch OUI data (%s)", strings.Join(errs, "; "))
}

func parseOUIJSON(body []byte) (OUIDatabase, error) {
	// First, try simple map form (prefix -> vendor).
	var simple OUIDatabase
	if err := json.Unmarshal(body, &simple); err == nil && len(simple) > 0 {
		return simple, nil
	}

	// Fallback: parse list of assignments (used by regauth.standards.ieee.org).
	var list []map[string]any
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	db := make(OUIDatabase)
	for _, item := range list {
		rawPrefix, _ := item["assignment"].(string)
		if rawPrefix == "" {
			continue
		}
		vendor, _ := item["organizationName"].(string)
		if vendor == "" {
			vendor, _ = item["org"].(string)
		}
		if vendor == "" {
			continue
		}
		prefix := strings.ToUpper(strings.ReplaceAll(rawPrefix, "-", ":"))
		db[prefix] = vendor
	}
	if len(db) == 0 {
		return nil, fmt.Errorf("no OUI entries decoded")
	}
	return db, nil
}

func parseOUICSV(body []byte) (OUIDatabase, error) {
	r := csv.NewReader(strings.NewReader(string(body)))
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("no csv rows")
	}

	// Detect header to find assignment/vendor columns.
	assignIdx, vendorIdx := -1, -1
	header := rows[0]
	for i, h := range header {
		switch strings.ToLower(strings.TrimSpace(h)) {
		case "assignment", "hex", "prefix":
			assignIdx = i
		case "organization name", "org", "vendor", "organization":
			vendorIdx = i
		}
	}

	start := 0
	if assignIdx >= 0 && vendorIdx >= 0 {
		start = 1 // skip header row
	} else {
		// assume first two columns
		assignIdx, vendorIdx = 0, 1
	}

	db := make(OUIDatabase)
	for _, row := range rows[start:] {
		if assignIdx >= len(row) || vendorIdx >= len(row) {
			continue
		}
		prefix := normalizePrefix(strings.TrimSpace(row[assignIdx]))
		vendor := strings.TrimSpace(row[vendorIdx])
		if prefix == "" || vendor == "" {
			continue
		}
		db[prefix] = vendor
	}
	if len(db) == 0 {
		return nil, fmt.Errorf("no OUI entries decoded from csv")
	}
	return db, nil
}

func parseOUITXT(body []byte) (OUIDatabase, error) {
	lines := strings.Split(string(body), "\n")
	db := make(OUIDatabase)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "OUI") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		prefix := normalizePrefix(fields[0])
		if prefix == "" {
			continue
		}
		vendor := strings.TrimSpace(strings.Join(fields[1:], " "))
		if vendor == "" {
			continue
		}
		db[prefix] = vendor
	}
	if len(db) == 0 {
		return nil, fmt.Errorf("no OUI entries decoded from txt")
	}
	return db, nil
}

func parseWiresharkManuf(body []byte) (OUIDatabase, error) {
	lines := strings.Split(string(body), "\n")
	db := make(OUIDatabase)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		prefixRaw := fields[0]
		vendor := strings.TrimSpace(strings.Join(fields[1:], " "))
		if vendor == "" {
			continue
		}

		prefix := normalizePrefix(prefixRaw)
		if prefix == "" {
			continue
		}

		db[prefix] = vendor
	}

	if len(db) == 0 {
		return nil, fmt.Errorf("no OUI entries decoded from manuf format")
	}
	return db, nil
}

func normalizePrefix(raw string) string {
	normalized := strings.ToUpper(strings.ReplaceAll(raw, "-", ":"))
	normalized = strings.ReplaceAll(normalized, ".", ":")
	normalized = strings.ReplaceAll(normalized, "/", "")
	if strings.Count(normalized, ":") == 0 && len(normalized) == 6 {
		// e.g. "A1B2C3" -> "A1:B2:C3"
		normalized = strings.ToUpper(raw)
		normalized = normalized[:2] + ":" + normalized[2:4] + ":" + normalized[4:6]
	}
	if len(normalized) != 8 || strings.Count(normalized, ":") != 2 {
		return ""
	}
	return normalized
}

func DiffOUI(local, master OUIDatabase) OUIDiff {
	diff := OUIDiff{
		TotalLocal:  len(local),
		TotalMaster: len(master),
	}

	for prefix, vendor := range master {
		if localVendor, exists := local[prefix]; !exists {
			diff.NewVendors = append(diff.NewVendors, prefix)
		} else if localVendor != vendor {
			diff.UpdatedVendors = append(diff.UpdatedVendors, prefix)
		}
	}

	for prefix := range local {
		if _, exists := master[prefix]; !exists {
			diff.RemovedVendors = append(diff.RemovedVendors, prefix)
		}
	}

	return diff
}

func ApplyOUIUpdate(local, master OUIDatabase) OUIDatabase {
	merged := make(OUIDatabase)

	for prefix, vendor := range master {
		merged[prefix] = vendor
	}

	return merged
}

func SaveLocalOUI(db OUIDatabase) error {
	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return err
	}

	// Try to save next to executable first
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		assetsDir := filepath.Join(exeDir, "assets")
		if err := os.MkdirAll(assetsDir, 0755); err == nil {
			diskPath := filepath.Join(assetsDir, "oui.json")
			if err := os.WriteFile(diskPath, data, 0644); err == nil {
				return nil
			}
		}
	}

	// Fallback to current directory (for development)
	return os.WriteFile("assets/oui.json", data, 0644)
}
