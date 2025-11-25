package storage

import "sync"

// Global app-level lock for any disk writes (JSON, history, vendor DB)
var FileLock = &sync.Mutex{}