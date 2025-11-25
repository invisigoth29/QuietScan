package vendor

import (
    "encoding/csv"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "net/http"
    "os"
    "strings"
    "time"

    "quietscan/storage"
)

const (
    ieeeURL      = "https://standards-oui.ieee.org/oui/oui.csv"
    localOuiFile = "assets/oui.json"
)

type VendorDiff struct {
    NewVendors     map[string]string
    UpdatedVendors map[string]string
    RemovedVendors map[string]string
    TotalMaster    int
    TotalLocal     int
}

func LoadLocalOUI() (map[string]string, error) {
    b, err := os.ReadFile(localOuiFile)
    if err != nil {
        return nil, err
    }
    var data map[string]string
    err = json.Unmarshal(b, &data)
    return data, err
}

func SaveLocalOUI(m map[string]string) error {
    storage.FileLock.Lock()
    defer storage.FileLock.Unlock()

    b, err := json.MarshalIndent(m, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(localOuiFile, b, 0644)
}

func FetchMasterOUI() (map[string]string, error) {
    client := &http.Client{Timeout: 20 * time.Second}
    resp, err := client.Get(ieeeURL)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        return nil, fmt.Errorf("unexpected HTTP status: %s", resp.Status)
    }

    reader := csv.NewReader(resp.Body)
    reader.FieldsPerRecord = -1

    master := make(map[string]string)

    _, _ = reader.Read()

    for {
        record, err := reader.Read()
        if errors.Is(err, io.EOF) {
            break
        }
        if err != nil {
            continue
        }

        if len(record) < 3 {
            continue
        }

        rawPrefix := strings.TrimSpace(record[1])
        orgName := strings.TrimSpace(record[2])

        if rawPrefix == "" || orgName == "" {
            continue
        }

        clean := strings.ToUpper(strings.ReplaceAll(rawPrefix, "-", ":"))
        master[clean] = orgName
    }

    return master, nil
}

func DiffOUI(local, master map[string]string) VendorDiff {
    diff := VendorDiff{
        NewVendors:     map[string]string{},
        UpdatedVendors: map[string]string{},
        RemovedVendors: map[string]string{},
        TotalLocal:     len(local),
        TotalMaster:    len(master),
    }

    for k, masterVal := range master {
        localVal, exists := local[k]
        if !exists {
            diff.NewVendors[k] = masterVal
        } else if localVal != masterVal {
            diff.UpdatedVendors[k] = masterVal
        }
    }

    for k := range local {
        if _, exists := master[k]; !exists {
            diff.RemovedVendors[k] = local[k]
        }
    }

    return diff
}

func ApplyOUIUpdate(local, master map[string]string) map[string]string {
    merged := make(map[string]string)

    for k, v := range master {
        merged[k] = v
    }

    for k, v := range local {
        if _, exists := merged[k]; !exists {
            merged[k] = v
        }
    }

    return merged
}